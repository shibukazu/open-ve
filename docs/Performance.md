# Performance

This document describes the performance of Open-VE.
We evaluate the performance of Open-VE by conducting E2E tests using [LOCUST](https://locust.io/) as part of a CI/CD pipeline with GitHub Actions.

The test execution file can be found [here](../test/locust/monolithic.py)
The Terraform file for the test instance can be found [here](../terraform/modules/aws/load_test/main.tf)

## Test Conditions

- Conditions
  - Maximum Users: 1000 users
  - User increase rate: 20 users/second
  - Test duration: 100 seconds
- Target Environment (Open-VE Instance)
  - Execution type: ECS Fargate
  - vCPU: 2048 (2v CPU)
  - Memory: 4096 MB
  - Region: ap-northeast-1
- Client
  - Github Actions Hosted Runner

## Result

![response_times_(ms)_1735885856 103](https://github.com/user-attachments/assets/df439626-678d-48f3-8dc6-1c1cb708a007)

[OktaFGA の E2E レスポンスタイム(非公式)](https://dev.classmethod.jp/articles/tokyo-oktafga-openfga/)
