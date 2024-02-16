from dataclasses import dataclass
from cycler import V
import pandas as pd
import matplotlib.pyplot as plt
from fastapi import FastAPI
from pydantic import Field, BaseModel
from typing import List
import uuid
from fastapi.encoders import jsonable_encoder
import networkx as nx
import matplotlib


from io import BytesIO
from numpy import random
from starlette.responses import StreamingResponse
from bokeh.palettes import Spectral, TolRainbow
import seaborn as sns
from typing import Hashable
import matplotlib.dates as mdates
from schemas import Points, Bar, TimeSeries, Graph, Plot

X, Y = "x", "y"


@dataclass
class Service:
    def __post_init__(self) -> None:
        matplotlib.use("agg")
        sns.set_theme(style="whitegrid")
        self.palette = sns.color_palette("husl", 8)
        self.default_figsize = (19.20, 10.80)

    def _fig_to_bytes(self, fig) -> BytesIO:
        buf = BytesIO()
        fig.savefig(buf, format="jpeg")
        buf.seek(0)
        plt.close(fig)

        return buf

    def _get_fig_and_ax(self, input_: Plot):
        fig = plt.figure(figsize=self.default_figsize)
        ax = fig.gca()

        if input_.title is not None:
            plt.title(input_.title)
        if input_.ylabel is not None:
            plt.ylabel(input_.ylabel)
        if input_.xlabel is not None:
            plt.xlabel(input_.xlabel)

        return fig, ax

    def draw_points(self, points: Points) -> BytesIO:
        points_list = [
            jsonable_encoder(item, by_alias=True) for item in points.__root__
        ]

        df = pd.DataFrame(points_list)
        fig = plt.figure(figsize=self.default_figsize)
        ax = fig.gca()

        for color in df["color"].unique():
            colored_df = df[df["color"] == color]
            colored_df.plot(
                kind="scatter",
                x="lon",
                y="lat",
                grid=True,
                stacked=True,
                ax=ax,
                color=color,
                s=10,
                fig=fig,
            )

        return self._fig_to_bytes(fig)

    def draw_bar(self, input_: Bar) -> BytesIO:
        fig, ax = self._get_fig_and_ax(input_)

        df = pd.DataFrame(input_.values.items(), columns=[X, Y])
        df = df.sort_values(by=Y, ascending=False)
        if input_.limit is not None:
            df = df.head(input_.limit)

        my_plot = sns.barplot(data=df, ax=ax, palette=self.palette, x=X, y=Y)
        my_plot.set(yticklabels=[])
        my_plot.set_xticklabels(
            my_plot.get_xticklabels(), rotation=45, horizontalalignment="right"
        )

        # add value above each bar
        for p in my_plot.patches:
            my_plot.annotate(
                format(p.get_height(), ".0f"),
                (p.get_x() + p.get_width() / 2.0, p.get_height()),
                ha="center",
                va="center",
                xytext=(0, 10),
                textcoords="offset points",
            )

        return self._fig_to_bytes(fig)

    def draw_timeseries(self, input_: TimeSeries) -> BytesIO:
        fig, ax = self._get_fig_and_ax(input_)

        df = pd.DataFrame(input_.values.items(), columns=[X, Y])

        sns.lineplot(data=df, ax=ax, palette=self.palette, x=X, y=Y)

        if input_.only_time:
            ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
            ax.xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))

        return self._fig_to_bytes(fig)

    def draw_graph(self, input_: Graph) -> BytesIO:
        fig, ax = self._get_fig_and_ax(input_)

        g = nx.DiGraph(directed=True)

        nodes = set()
        avg = 0.0
        edgewidths = []
        for edge in input_.edges:
            g.add_edge(edge.first, edge.second, weight=edge.weight)
            avg += edge.weight
            nodes.add(edge.first)
            nodes.add(edge.second)
            edgewidths.append(edge.weight)

        avg = avg / len(input_.edges)
        max_ = max(edgewidths)

        for idx in range(len(edgewidths)):
            edgewidths[idx] = edgewidths[idx] / max_

        # positions for all nodes - seed for reproducibility
        pos = nx.circular_layout(g)

        TolRainbow23 = TolRainbow[23]
        TolRainbow23_count, TolRainbow23_mod = divmod(len(nodes), 23)
        pallete = TolRainbow23 * TolRainbow23_count + TolRainbow23[:TolRainbow23_mod]

        # nodes
        nx.draw_networkx_nodes(g, pos, node_size=1000, ax=ax, node_color=pallete)

        # edges
        nx.draw_networkx_edges(g, pos, width=3, alpha=edgewidths, ax=ax)

        # node labels
        nx.draw_networkx_labels(
            g,
            pos,
            font_size=10,
            font_family="sans-serif",
            ax=ax,
            bbox={"ec": "k", "fc": "white", "alpha": 1},
        )
        ax.margins(0.008)
        ax.collections[0].set_edgecolor("#000000")

        return self._fig_to_bytes(fig)
