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
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:DescribeRepositories",
      "ecr:ListTagsForResource",
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:CreateRepository", // 追加
      "ecr:DeleteRepository", // 追加

      // ECS
      "ecs:RunTask",
      "ecs:StopTask",
      "ecs:DescribeTasks",
      "ecs:ListTasks",
      "ecs:DescribeClusters",
      "ecs:DescribeTaskDefinition",
      "ecs:DescribeServices",
      "ecs:CreateCluster", // 追加
      "ecs:DeleteCluster", // 追加
      "ecs:RegisterTaskDefinition",
      "ecs:DeregisterTaskDefinition",
      "ecs:CreateService", // 追加
      "ecs:DeleteService", // 追加
      "ecs:UpdateService", // 追加

      // EC2 (VPC関連)
      "ec2:DescribeVpcs",
      "ec2:DescribeAvailabilityZones",
      "ec2:DescribeVpcAttribute",
      "ec2:DescribeSubnets",
      "ec2:DescribeSecurityGroups",
      "ec2:CreateVpc",             // 追加
      "ec2:DeleteVpc",             // 追加
      "ec2:CreateSubnet",          // 追加
      "ec2:DeleteSubnet",          // 追加
      "ec2:CreateInternetGateway", // 追加
      "ec2:AttachInternetGateway", // 追加
      "ec2:DeleteInternetGateway", // 追加
      "ec2:CreateRouteTable",      // 追加
      "ec2:DeleteRouteTable",      // 追加
      "ec2:CreateRoute",           // 追加
      "ec2:AssociateRouteTable",   // 追加

      // IAM
      "iam:GetRole",
      "iam:ListRolePolicies",
      "iam:ListAttachedRolePolicies",
      "iam:PassRole",
      "iam:CreateRole", // 追加
      "iam:DeleteRole", // 追加

      // Security Groups
      "ec2:CreateSecurityGroup",           // 追加
      "ec2:DeleteSecurityGroup",           // 追加
      "ec2:AuthorizeSecurityGroupIngress", // 追加
      "ec2:RevokeSecurityGroupIngress",    // 追加
      "ec2:AuthorizeSecurityGroupEgress",  // 追加
      "ec2:RevokeSecurityGroupEgress"      // 追加
    ]

    resources = ["*"]
  }
}
