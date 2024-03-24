from pydantic import Field, BaseModel


class AnimePredictionResponse(BaseModel):
    prediction: float = Field(..., example=0.7)
