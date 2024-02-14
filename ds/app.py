import pandas as pd
import matplotlib.pyplot as plt
from fastapi import FastAPI
from pydantic import Field, BaseModel
from typing import List
import uuid
from fastapi.encoders import jsonable_encoder


from io import BytesIO
from numpy import random
from starlette.responses import StreamingResponse
from service import Service
from schemas import Points, Bar


app = FastAPI()
service = Service()


@app.get("/health")
def health() -> dict[str, bool]:
    return {"success": True}


@app.post("/draw-points", response_class=StreamingResponse)
def draw_fig(points: Points) -> StreamingResponse:
    return StreamingResponse(service.draw_points(points), media_type="image/jpeg")


@app.post("/draw-histogram", response_class=StreamingResponse)
def draw_histogram(input_: Bar) -> StreamingResponse:
    return StreamingResponse(service.draw_bar(input_), media_type="image/jpeg")
