from dataclasses import dataclass
import pandas as pd
import matplotlib.pyplot as plt
from fastapi.encoders import jsonable_encoder
import networkx as nx
import matplotlib

from datetime import datetime

from io import BytesIO
import seaborn as sns
import matplotlib.dates as mdates
from schemas import Points, Bar, TimeSeries, Graph, Plot, GraphEdge
from matplotlib.lines import Line2D

X, Y = "x", "y"


@dataclass
class Service:
    def __post_init__(self) -> None:
        matplotlib.use("agg")
        sns.set_theme(style="whitegrid")
        self.default_figsize = (20, 10)

    def _get_palette(self, n: int):
        return sns.color_palette("Set2", n)

    def _fig_to_bytes(self, fig) -> BytesIO:
        buf = BytesIO()
        fig.savefig(
            buf,
            format="jpeg",
            dpi=300,
        )
        buf.seek(0)
        plt.close(fig)

        return buf

    def _get_fig_and_ax(self, input_: Plot):
        fig = plt.figure(figsize=self.default_figsize)
        ax = fig.gca()

        if input_.title is not None:
            ax.set_title(input_.title)
        if input_.ylabel is not None:
            ax.set_ylabel(input_.ylabel)
        if input_.xlabel is not None:
            ax.set_xlabel(input_.xlabel)

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

        my_plot = sns.barplot(
            data=df, ax=ax, palette=self._get_palette(len(input_.values)), x=X, y=Y
        )
        my_plot.set(yticklabels=[])
        my_plot.set_xticklabels(
            my_plot.get_xticklabels(), rotation=30, horizontalalignment="right"
        )

        # add value above each bar
        for p in my_plot.patches:
            height = p.get_height()
            my_plot.annotate(
                (
                    format(height, ".0f")
                    if height == round(height, 2)
                    else format(height, ".2f")
                ),
                (p.get_x() + p.get_width() / 2.0, height),
                ha="center",
                va="center",
                xytext=(0, 10),
                textcoords="offset points",
            )

        return self._fig_to_bytes(fig)

    def draw_timeseries(self, input_: TimeSeries) -> BytesIO:
        fig, ax = self._get_fig_and_ax(input_)

        # make table
        table: list[tuple[str, datetime, float]] = []
        for legend, values in input_.values.items():
            row: list[tuple[str, datetime, float]] = []
            for date, value in values.items():
                row.append((legend, date, value))
            table.extend(row)

        df = pd.DataFrame(table, columns=["legend", "date", "value"])
        df = df.drop_duplicates(subset=["legend", "date", "value"])
        df_wide = df.pivot_table(
            index="date", columns="legend", values="value", aggfunc="sum"
        )

        sns.lineplot(
            data=df_wide,
            ax=ax,
            palette=self._get_palette(len(input_.values)),
            marker="o",
            linestyle=(0, (1, 10)),
        )

        if input_.only_time:
            ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
            ax.xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))

        return self._fig_to_bytes(fig)

    def concat_graph(self, input_: Graph) -> None:
        one_to_other: dict[str, dict[str, float]] = {}
        for edge in input_.edges:
            v = one_to_other.get(edge.first, None)
            if v is None:
                one_to_other[edge.first] = {edge.second: edge.weight}
                continue

            point = v.get(edge.second, None)
            if point is None:
                one_to_other[edge.first][edge.second] = edge.weight
                continue

            one_to_other[edge.first][edge.second] += edge.weight

        new_input: list[GraphEdge] = []
        for first, edge in one_to_other.items():
            for second, v in edge.items():
                new_input.append(GraphEdge(first=first, second=second, weight=v))

        input_.edges = new_input

    def draw_graph_as_heatmap(self, input_: Graph) -> BytesIO:
        fig, ax = self._get_fig_and_ax(input_)

        items: dict[str, dict[str, float]] = {}
        for edge in input_.edges:
            ok = items.get(edge.first)
            if ok is None:
                items[edge.first] = {edge.second: edge.weight}
                continue

            items[edge.first][edge.second] = edge.weight

        df = pd.DataFrame(
            [
                {"first": first} | {second: weight for second, weight in item.items()}
                for first, item in items.items()
            ],
        )
        df = df.set_index("first")
        index = df.index.union(df.columns)
        df = df.reindex(index=index, columns=index)

        plot = sns.heatmap(data=df, ax=ax, linewidth=0.5, cmap=sns.cm.rocket_r)
        plot.set_xticklabels(
            plot.get_xticklabels(), rotation=30, horizontalalignment="right"
        )
        if input_.ylabel is not None:
            plot.set_ylabel(input_.ylabel)
        if input_.xlabel is not None:
            plot.set_xlabel(input_.xlabel)

        return self._fig_to_bytes(fig)

    def draw_graph(self, input_: Graph) -> BytesIO:
        self.concat_graph(input_)

        fig, ax = self._get_fig_and_ax(input_)

        g = nx.Graph()

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

        colors = [
            "#bfdbf7",  # light blue
            "#053c5e",  # dark blue
            "#a31621",  # red
        ]
        max_ = max(edgewidths)

        for idx in range(len(edgewidths)):
            edgewidths[idx] = edgewidths[idx] / max_
            if edgewidths[idx] < 0.7:
                if edgewidths[idx] < 0.3:
                    edgewidths[idx] = colors[0]
                    continue
                edgewidths[idx] = colors[1]
                continue
            edgewidths[idx] = colors[2]

        # positions for all nodes - seed for reproducibility
        pos = nx.circular_layout(g)

        # nodes
        nx.draw_networkx_nodes(
            g,
            pos,
            node_size=3000,
            ax=ax,
            node_color=sns.color_palette("Set2", len(nodes)),
        )

        # edges
        nx.draw_networkx_edges(
            g,
            pos,
            width=3,
            alpha=1,
            ax=ax,
            edge_color=edgewidths,
        )

        # node labels
        nx.draw_networkx_labels(
            g,
            pos,
            font_size=15,
            font_family="sans-serif",
            ax=ax,
            bbox={"ec": "k", "fc": "white", "alpha": 1},
        )

        proxies = [Line2D([0, 1], [0, 1], color=color, lw=7) for color in colors]
        labels = ["<30% percents", "30%-70% percents", ">70% percents"]
        ax.legend(proxies, labels)

        return self._fig_to_bytes(fig)
