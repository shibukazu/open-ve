from typing import Any
from urllib.parse import urljoin

import requests
from locust import HttpUser, events, task
from locust.env import Environment
from utils.requests import (
    REGISTER_ENDPOINT,
    VALIDATE_ENDPOINT,
    RegisterRequestBody,
    RegisterRequestValidation,
    RegisterRequestVariable,
    ValidateRequestBody,
    ValidateRequestValidation,
)


@events.test_start.add_listener
def on_test_start(environment: Environment, **kwargs: Any) -> None:
    if environment.host is None:
        raise Exception("Host is not set")

    body = RegisterRequestBody(
        validations=[
            RegisterRequestValidation(
                id="item",
                cels=[
                    "price > 0",
                    "size(image) < 360",
                ],
                variables=[
                    RegisterRequestVariable(
                        name="price",
                        type="int",
                    ),
                    RegisterRequestVariable(
                        name="image",
                        type="bytes",
                    ),
                ],
            ),
            RegisterRequestValidation(
                id="user",
                cels=[
                    "size(name) < 20",
                ],
                variables=[
                    RegisterRequestVariable(
                        name="name",
                        type="string",
                    ),
                ],
            ),
        ]
    )

    print(body.model_dump_json())

    res = requests.post(
        urljoin(environment.host, REGISTER_ENDPOINT),
        json=body.model_dump(),
    )

    if res.status_code != 200:
        print(f"Failed to register request: {res.text}")
        exit(1)


class User(HttpUser):
    @task
    def validate(self) -> None:
        body = ValidateRequestBody(
            validations=[
                ValidateRequestValidation(
                    id="item",
                    variables={
                        "price": -100,
                        "image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGO4unY2AAR4Ah51j5XwAAAAAElFTkSuQmCC",
                    },
                ),
                ValidateRequestValidation(
                    id="user",
                    variables={
                        "name": "longlonglonglongname",
                    },
                ),
            ]
        )
        with self.client.post(
            VALIDATE_ENDPOINT,
            json=body.model_dump(),
            catch_response=True,
        ) as res:
            if res.status_code != 200:
                res.failure(f"Failed to validate request: {res.text}")
