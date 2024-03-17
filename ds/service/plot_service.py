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
from schemas.plot import Points, Bar, TimeSeries, Graph, Plot, GraphEdge
import math
from networkx.drawing.nx_pydot import graphviz_layout
from matplotlib.lines import Line2D
from schemas.plot import GraphLayout

X, Y = "x", "y"


@dataclass
class PlotService:
    def __post_init__(self) -> None:
        matplotlib.use("agg")
        sns.set_theme(style="whitegrid")

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
        fig = plt.figure(figsize=input_.figsize)
        ax = fig.gca()

        if input_.title is not None:
            ax.set_title(input_.title)
        if input_.ylabel is not None:
            ax.set_ylabel(input_.ylabel)
        if input_.xlabel is not None:
            ax.set_xlabel(input_.xlabel)

        return fig, ax

    def draw_points(self, points: Points) -> BytesIO:
        points_list = [jsonable_encoder(item, by_alias=True) for item in points.points]

        df = pd.DataFrame(points_list)
        fig = plt.figure(figsize=points.figsize)
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
        df = df.sort_values(by=Y, ascending=input_.asc)
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

    def _prepare_graph(self, input_: Graph) -> None:
        for edge in input_.edges:
            edge.first = edge.first.replace(":", "_")
            edge.second = edge.second.replace(":", "_")

        if input_.nodes is not None:
            new_nodes = {}
            for k, v in input_.nodes.items():
                new_nodes[k.replace(":", "_")] = v

            input_.nodes = new_nodes

        if input_.root_node is not None:
            input_.root_node = input_.root_node.replace(":", "_")

    def _normalize(self, list_: list[float], start: float, end: float) -> None:
        max_ = max(list_)
        min_ = min(list_)
        dif = max_ - min_
        if dif == 0:
            return None

        start_end_dif = end - start

        for i in range(len(list_)):
            list_[i] = ((list_[i] - min_) / dif * start_end_dif) + start

    def draw_graph(self, input_: Graph) -> BytesIO:
        self.concat_graph(input_)
        self._prepare_graph(input_)

        fig, ax = self._get_fig_and_ax(input_)

        g = nx.Graph()

        nodes = set()
        avg = 0.0
        edgewidths = []
        for edge in input_.edges:
            edge.first = edge.first
            edge.second = edge.second
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

        if max_ == 0:
            for idx in range(len(edgewidths)):
                edgewidths[idx] = colors[1]
        else:
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
        if input_.layout == GraphLayout.CIRCULAR_LAYOUT:
            pos = nx.circular_layout(g)
        elif input_.layout == GraphLayout.SPRING_LAYOUT:
            pos = nx.spring_layout(g, k=5 / math.sqrt(g.order()))
        elif input_.layout == GraphLayout.SPECTRAL_LAYOUT:
            pos = nx.spectral_layout(g)
        elif input_.layout == GraphLayout.CIRCULAR_TREE_LAYOUT:
            pos = graphviz_layout(g, prog="twopi", root=input_.root_node)
        else:
            pos = nx.circular_layout(g)

        if input_.nodes is None:
            # nodes
            nx.draw_networkx_nodes(
                g,
                pos,
                node_size=3000,
                ax=ax,
                node_color=sns.color_palette("Set2", len(nodes)),
            )
        else:
            nodelist: list[str] = []
            nodeweign: list[float] = []
            for nodename, node in input_.nodes.items():
                nodelist.append(nodename)
                nodeweign.append(node.weight)

            self._normalize(nodeweign, start=1000, end=5000)

            nx.draw_networkx_nodes(
                g,
                pos,
                node_size=nodeweign,
                ax=ax,
                node_color=sns.color_palette("Set2", len(nodes)),
                nodelist=nodelist,
            )

        # edges
        nx.draw_networkx_edges(
            g,
            pos,
            width=3,
            alpha=1,
            ax=ax,
            edge_color=edgewidths if input_.weighted_edges else "k",
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

        if input_.weighted_edges:
            proxies = [Line2D([0, 1], [0, 1], color=color, lw=7) for color in colors]
            labels = ["<30%", "30%-70%", ">70%"]
            ax.legend(proxies, labels)

        return self._fig_to_bytes(fig)
