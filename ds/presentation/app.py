from fastapi import FastAPI


from starlette.responses import StreamingResponse
from schemas.plot import Points, Bar, TimeSeries, Graph
from shared.container import init_combat_container
from schemas.anime import AnimePredictionResponse
from fastapi import Request
import time

from fastapi import UploadFile


app = FastAPI()
container = init_combat_container()


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
    return StreamingResponse(
        container.plot_service.draw_points(points), media_type="image/jpeg"
    )


@app.post("/histogram", response_class=StreamingResponse)
def draw_bar(input_: Bar) -> StreamingResponse:
    return StreamingResponse(
        container.plot_service.draw_bar(input_), media_type="image/jpeg"
    )


@app.post("/timeseries", response_class=StreamingResponse)
def draw_timeseries(input_: TimeSeries) -> StreamingResponse:
    return StreamingResponse(
        container.plot_service.draw_timeseries(input_), media_type="image/jpeg"
    )


@app.post("/graph", response_class=StreamingResponse)
def draw_graph(input_: Graph) -> StreamingResponse:
    return StreamingResponse(
        container.plot_service.draw_graph(input_), media_type="image/jpeg"
    )


@app.post("/graph-as-heatmap", response_class=StreamingResponse)
def draw_graph_as_heatmap(input_: Graph) -> StreamingResponse:
    return StreamingResponse(
        container.plot_service.draw_graph_as_heatmap(input_), media_type="image/jpeg"
    )


@app.post("/anime/predict", response_model=AnimePredictionResponse)
async def amine_predict(image: UploadFile) -> AnimePredictionResponse:
    prediction = container.anime_service.make_prediction(await image.read())
    return AnimePredictionResponse(prediction=prediction)
