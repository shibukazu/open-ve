locals {
  prefix = "open-ve_oidc"
}

resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com"
  ]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
  ]
}



resource "aws_iam_role" "github_oidc_role" {
  name = "${local.prefix}-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity",
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        },
        Condition = {
          StringLike = {
            "token.actions.githubusercontent.com:sub" : [
              "repo:shibukazu/open-ve:*",
            ]
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "github_oidc_policy" {
  name   = "${local.prefix}-policy"
  role   = aws_iam_role.github_oidc_role.id
  policy = data.aws_iam_policy_document.github_oidc_access_policy.json
}

data "aws_iam_policy_document" "github_oidc_access_policy" {
  statement {
    effect = "Allow"

    actions = [
      // ECR
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      // ECS
      "ecs:RunTask",
      "ecs:StopTask",
      "ecs:DescribeTasks",
      "ecs:RegisterTaskDefinition",
      "ecs:DeregisterTaskDefinition",
      // S3
      "s3:GetObject",
      "s3:PutObject",
      // EC2(VPC)
      "ec2:DescribeNetworkInterfaces",
      // IAM
      "iam:PassRole",
    ]

    resources = ["*"]
  }
}
