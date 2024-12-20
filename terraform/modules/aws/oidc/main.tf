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

data "aws_iam_policy_document" "github_oidc_trust" {
  statement {
    effect = "Allow"

    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
        "repo:shibukazu/open-ve:ref:refs/heads/main"
      ]
    }
  }
}

resource "aws_iam_role" "github_oidc_role" {
  name               = "${local.prefix}-role"
  assume_role_policy = data.aws_iam_policy_document.github_oidc_trust.json
}

resource "aws_iam_role_policy" "github_oidc_policy" {
  role   = aws_iam_role.github_oidc_role.id
  policy = data.aws_iam_policy_document.github_oidc_access_policy.json
}

data "aws_iam_policy_document" "github_oidc_access_policy" {
  statement {
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:PutImage",
      "ecs:RunTask",
      "ecs:StopTask",
      "ecs:DescribeTasks",
      "ecs:ListTasks"
    ]

    resources = ["*"]
  }
}

output "oidc_role_arn" {
  value = aws_iam_role.github_oidc_role.arn
}