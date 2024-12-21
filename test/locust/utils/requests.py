from typing import Any

from pydantic import BaseModel

REGISTER_ENDPOINT = "/v1/dsl"


class RegisterRequestVariable(BaseModel):
    name: str
    type: str


class RegisterRequestValidation(BaseModel):
    id: str
    cels: list[str]
    variables: list[RegisterRequestVariable]


class RegisterRequestBody(BaseModel):
    validations: list[RegisterRequestValidation]


VALIDATE_ENDPOINT = "/v1/check"


class ValidateRequestValidation(BaseModel):
    id: str
    variables: dict[str, Any]


class ValidateRequestBody(BaseModel):
    validations: list[ValidateRequestValidation]
