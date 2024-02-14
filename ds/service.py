from dataclasses import dataclass
from cycler import V
import pandas as pd
import matplotlib.pyplot as plt
from fastapi import FastAPI
from pydantic import Field, BaseModel
from typing import List
import uuid
from fastapi.encoders import jsonable_encoder
import matplotlib


from io import BytesIO
from numpy import random
from starlette.responses import StreamingResponse
import seaborn as sns
from typing import Hashable
from schemas import Points, Bar


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
        fig = plt.figure(figsize=self.default_figsize)
        ax = fig.gca()

        df = pd.DataFrame(input_.values.items(), columns=["x", "y"])
        df = df.sort_values(by="y", ascending=False)
        if input_.limit is not None:
            df = df.head(input_.limit)

        my_plot = sns.barplot(data=df, ax=ax, palette=self.palette, x="x", y="y")
        my_plot.set(yticklabels=[])
        my_plot.set_xticklabels(
            my_plot.get_xticklabels(), rotation=45, horizontalalignment="right"
        )

        plt.title(input_.title)
        plt.ylabel(input_.ylabel)
        plt.xlabel(input_.xlabel)
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
