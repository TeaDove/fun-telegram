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
from schemas import Points, Bar, TimeSeries, Graph
from fastapi import Request
import time
from loguru import logger

app = FastAPI()
service = Service()


@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    t0 = time.monotonic()
    response = await call_next(request)
    process_time = time.monotonic() - t0
    response.headers["X-Process-Time"] = str(process_time)
    return response


@app.get("/health")
def health() -> dict[str, bool]:
    return {"success": True}


@app.post("/points", response_class=StreamingResponse)
def draw_fig(points: Points) -> StreamingResponse:
    return StreamingResponse(service.draw_points(points), media_type="image/jpeg")


@app.post("/histogram", response_class=StreamingResponse)
def draw_bar(input_: Bar) -> StreamingResponse:
    return StreamingResponse(service.draw_bar(input_), media_type="image/jpeg")


@app.post("/timeseries", response_class=StreamingResponse)
def draw_timeseries(input_: TimeSeries) -> StreamingResponse:
    return StreamingResponse(service.draw_timeseries(input_), media_type="image/jpeg")


@app.post("/graph", response_class=StreamingResponse)
def draw_graph(input_: Graph) -> StreamingResponse:
    return StreamingResponse(service.draw_graph(input_), media_type="image/jpeg")


@app.post("/graph-as-heatmap", response_class=StreamingResponse)
def draw_graph_as_heatmap(input_: Graph) -> StreamingResponse:
    return StreamingResponse(
        service.draw_graph_as_heatmap(input_), media_type="image/jpeg"
    )
