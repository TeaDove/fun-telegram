from pydantic import Field, BaseModel
import uuid


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


class Bar(BaseModel):
    values: dict[str, float] = Field(
        example={
            random.choice(_random_names): random.randint(0, 100)
            for _ in range(random.randint(10, 20))
        }
    )

    title: str = "Title"
    ylabel: str = "Y Label"
    xlabel: str = "X Label"
    limit: int | None = None
