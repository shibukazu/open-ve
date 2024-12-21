locals {
  prefix = "open-ve_load_test"
}

// AWS ECR

resource "aws_ecr_repository" "repo" {
  name         = "${local.prefix}-repo"
  force_delete = true
}


// AWS ECS

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.prefix}-igw"
  }
}


resource "aws_vpc" "vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Name = "${local.prefix}-vpc"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "${local.prefix}-public-rt"
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "public_subnet" {
  count                   = 1
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = cidrsubnet("10.0.0.0/16", 4, count.index)
  map_public_ip_on_launch = true
  availability_zone       = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${local.prefix}-public-subnet-${count.index}"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public_subnet[0].id
  route_table_id = aws_route_table.public.id
}

resource "aws_ecs_cluster" "cluster" {
  name = "${local.prefix}-ecs_cluster"
}

resource "aws_ecs_task_definition" "task" {
  family                   = "${local.prefix}-ecs_task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "1024"

  container_definitions = jsonencode([
    {
      name      = "${local.prefix}-container"
      image     = aws_ecr_repository.repo.repository_url
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        },
        {
          containerPort = 9000
          hostPort      = 9000
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "OPEN-VE_MODE"
          value = "master"
        },
        {
          name  = "OPEN-VE_AUTHN_METHOD"
          value = "preshared"
        },
        {
          name  = "OPEN-VE_AUTHN_PRESHARED_KEY"
          value = var.preshared_key
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-region        = "ap-northeast-1" # 使用しているリージョン
          awslogs-group         = "/ecs/${local.prefix}-logs"
          awslogs-stream-prefix = "ecs"
        }
      }
      healthCheck = {
        command     = ["CMD-SHELL", "STATUS=$(curl -s http://localhost:8080/healthz | jq -r .status); if [ \"$STATUS\" == \"SERVING\" ]; then exit 0; else exit 1; fi"]
        interval    = 5
        timeout     = 5
        retries     = 3
        startPeriod = 10
      }
    }
  ])

  execution_role_arn = aws_iam_role.execution_role.arn
  task_role_arn      = aws_iam_role.task_role.arn
}

resource "aws_iam_role" "execution_role" {
  name = "${local.prefix}-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
  ]
}

resource "aws_iam_role" "task_role" {
  name = "ecsTaskRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_ecs_service" "service" {
  name            = "${local.prefix}-ecs_service"
  cluster         = aws_ecs_cluster.cluster.id
  task_definition = aws_ecs_task_definition.task.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.public_subnet[0].id]
    security_groups  = [aws_security_group.service_sg.id]
    assign_public_ip = true
  }
}
resource "aws_security_group" "service_sg" {
  name        = "${local.prefix}-service-security-group"
  description = "Allow service traffic"
  vpc_id      = aws_vpc.vpc.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 9000
    to_port     = 9000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "ecr_repository_url" {
  value = aws_ecr_repository.repo.repository_url
}

output "public_subnet" {
  value = aws_subnet.public_subnet[0].id
}

output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.cluster.name
}

output "ecs_task_definition" {
  value = aws_ecs_task_definition.task.family
}

output "ecs_service_name" {
  value = aws_ecs_service.service.name
}

output "service_security_group_id" {
  value = aws_security_group.service_sg.id
}
