import pandas as pd
import matplotlib.pyplot as plt
from fastapi import FastAPI
from pydantic import Field, BaseModel
from typing import List
import uuid
from fastapi.encoders import jsonable_encoder
import uvicorn


from io import BytesIO
from numpy import random
from starlette.responses import StreamingResponse
import matplotlib

matplotlib.use("agg")
colors = ("RED", "BLUE", "GREEN")
app = FastAPI()


class Point(BaseModel):
    id_: uuid.UUID = Field(default_factory=uuid.uuid4, alias="id")
    color: str

    lat: float
    lon: float


class Points(BaseModel):
    __root__: List[Point] = Field(
        example=[
            Point(
                color=random.choice(colors),
                lat=random.rand() * 20 + 40,
                lon=random.rand() * 20 + 40,
            )
            for _ in range(random.randint(10, 20))
        ]
    )


@app.post("/draw-points", response_class=StreamingResponse)
def draw_fig(points: Points) -> StreamingResponse:
    points_list = [jsonable_encoder(item, by_alias=True) for item in points.__root__]
    df = pd.DataFrame(points_list)
    fig = plt.figure()
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

    buf = BytesIO()
    fig.savefig(buf, format="png")
    buf.seek(0)
    plt.close(fig)
    return StreamingResponse(buf, media_type="image/png")


if __name__ == "__main__":
    uvicorn.run("draw:app", host="0.0.0.0", port=8000)
