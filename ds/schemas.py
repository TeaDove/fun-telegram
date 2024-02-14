from pydantic import Field, BaseModel
import uuid

from datetime import datetime
from examples import adj_matrix
from numpy import random

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


class Point(BaseModel):
    id_: uuid.UUID = Field(default_factory=uuid.uuid4, alias="id")
    color: str

    lat: float
    lon: float


class Points(BaseModel):
    __root__: list[Point] = Field(
        example=[
            Point(
                color=random.choice(colors),
                lat=random.rand() * 20 + 40,
                lon=random.rand() * 20 + 40,
            )
            for _ in range(random.randint(10, 20))
        ]
    )


class Plot(BaseModel):
    title: str = "Title"
    ylabel: str = "Y Label"
    xlabel: str = "X Label"


class Bar(Plot):
    values: dict[str, float] = Field(
        example={
            random.choice(_random_names): random.randint(0, 100)
            for _ in range(random.randint(10, 20))
        }
    )

    limit: int | None = None


class TimeSeries(Bar):
    values: dict[datetime, float] = Field(
        example={
            datetime(2023, 1, 1): random.randint(0, 100),
            datetime(2023, 2, 1): random.randint(0, 100),
            datetime(2023, 3, 1): random.randint(0, 100),
            datetime(2023, 4, 1): random.randint(0, 100),
            datetime(2023, 5, 1): random.randint(0, 100),
        }
    )

    only_time: bool = False


class Graph(Bar):
    values: list[list[int]] = Field(example=adj_matrix)

    names: list[str] = Field(example=[random.choice(_random_names) for _ in range(50)])
