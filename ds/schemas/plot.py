from pydantic import Field, BaseModel
import uuid

from datetime import datetime
from numpy import random
import enum


colors = ("RED", "BLUE", "GREEN")
_random_names = [
    "maggot",
    "varam",
    "kammus",
    "sarayn",
    "enarvyne",
    "alammu",
    "irethi",
    "neldam",
    "dren",
    "anasour",
    "irarvy",
    "vandal",
    "tadaves",
    "seran",
    "llaalam",
    "worker",
    "dalamus",
    "vandal",
    "gidave",
    "sendal",
    "othralen",
    "tedril",
    "girothran",
    "ararvy",
    "maryon",
    "llaala",
    "faveran",
    "gadaves",
    "uradasou",
    "berendal",
    "maggot",
    "heloth",
    "neldammu",
    "othren",
    "midaves",
    "deras",
    "vedaves",
    "ienevala",
]


class Plot(BaseModel):
    title: str | None = None
    ylabel: str | None = None
    xlabel: str | None = None
    figsize: tuple[int, int] = (20, 13)


class Point(BaseModel):
    id_: uuid.UUID = Field(default_factory=uuid.uuid4, alias="id")
    color: str

    lat: float
    lon: float


class Points(Plot):
    points: list[Point] = Field(
        example=[
            Point(
                color=random.choice(colors),
                lat=random.rand() * 20 + 40,
                lon=random.rand() * 20 + 40,
            )
            for _ in range(random.randint(10, 20))
        ]
    )


class Bar(Plot):
    values: dict[str, float] = Field(
        example={
            random.choice(_random_names): random.randint(0, 100)
            for _ in range(random.randint(10, 20))
        }
    )

    limit: int | None = Field(None, example=None)
    asc: bool = True


class TimeSeries(Plot):
    values: dict[str, dict[datetime, float]] = Field(
        example={
            random.choice(_random_names): {
                datetime(
                    2024, random.randint(1, 12), random.randint(1, 28)
                ): random.randint(0, 100)
                for _ in range(20)
            }
            for _ in range(3)
        }
    )

    only_time: bool = False


class GraphEdge(BaseModel):
    first: str
    second: str
    weight: float


class GraphNode(BaseModel):
    image: bytes | None = None
    weight: float = 1


class GraphLayout(str, enum.Enum):
    SPRING_LAYOUT = "spring"
    CIRCULAR_LAYOUT = "circular"
    SPECTRAL_LAYOUT = "spectral"
    CIRCULAR_TREE_LAYOUT = "circular_tree"


class Graph(Plot):
    edges: list[GraphEdge] = Field(
        example=[
            GraphEdge(
                first=random.choice(_random_names),
                second=random.choice(_random_names),
                weight=random.randint(0, 100),
            )
            for _ in range(30)
        ]
    )
    weighted_edges: bool = True
    layout: GraphLayout = GraphLayout.CIRCULAR_LAYOUT
    nodes: dict[str, GraphNode] | None = None
    root_node: str | None = None
