import uvicorn
import multiprocessing as mp


if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8000, workers=mp.cpu_count())
