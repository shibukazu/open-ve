from typing import Any
from urllib.parse import urljoin

import requests
from utils.requests import (
    REGISTER_ENDPOINT,
    VALIDATE_ENDPOINT,
    RegisterRequestBody,
    RegisterRequestValidation,
    RegisterRequestVariable,
    ValidateRequestBody,
    ValidateRequestValidation,
)

from locust import HttpUser, events, task
from locust.argument_parser import LocustArgumentParser
from locust.env import Environment


@events.init_command_line_parser.add_listener
def init_parser(parser: LocustArgumentParser) -> None:
    parser.add_argument(
        "--auth-token",
        type=str,
        required=True,
        help="Authentication token to be used in requests",
    )


@events.test_start.add_listener
def on_test_start(environment: Environment, **kwargs: Any) -> None:
    if environment.host is None:
        raise Exception("Host is not set")
    if (
        environment.parsed_options is None
        or environment.parsed_options.auth_token is None
    ):
        raise Exception("Auth token is not set")
    headers = {
        "Authorization": f"Bearer {environment.parsed_options.auth_token}",
    }

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
        headers=headers,
    )

    if res.status_code != 200:
        print(f"Failed to register request: {res.text}")
        exit(1)


class User(HttpUser):
    def on_start(self) -> None:
        self.headers = {
            "Authorization": f"Bearer {self.environment.parsed_options.auth_token}",
        }

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
            headers=self.headers,
            catch_response=True,
        ) as res:
            if res.status_code != 200:
                res.failure(f"Failed to validate request: {res.text}")
