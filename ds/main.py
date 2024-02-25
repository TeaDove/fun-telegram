import uvicorn
import multiprocessing as mp
from loguru import logger

if __name__ == "__main__":
    logger.info("app starting")
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=8000,
        workers=mp.cpu_count(),
        log_level="warning",
    )
