#!/bin/bash

# Variables
ROLE_NAME="svg-webp-app-role"
POLICY_NAME="svg-webp-app-policy"
ASSUME_ROLE_POLICY_DOCUMENT="assume-role-policy.json"
POLICY_DOCUMENT="inline-policy.json"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)
ROLE_ARN="arn:aws:iam::${AWS_ACCOUNT_ID}:role/${ROLE_NAME}"

# JSON for Assume Role Policy
cat > $ASSUME_ROLE_POLICY_DOCUMENT <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF

# JSON for Inline Policy
cat > $POLICY_DOCUMENT <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::your-bucket-name",
                "arn:aws:s3:::your-bucket-name/*"
            ]
        }
    ]
}
EOF

# Check if the role already exists
echo "Checking if IAM role $ROLE_NAME exists..."
aws iam get-role --role-name $ROLE_NAME &> /dev/null

if [ $? -ne 0 ]; then
    echo "Role $ROLE_NAME does not exist. Creating..."
    aws iam create-role \
        --role-name $ROLE_NAME \
        --assume-role-policy-document file://$ASSUME_ROLE_POLICY_DOCUMENT

    if [ $? -eq 0 ]; then
        echo "Role $ROLE_NAME created successfully."
    else
        echo "Failed to create role $ROLE_NAME. Exiting."
        exit 1
    fi
else
    echo "Role $ROLE_NAME already exists."
fi

# Check if the policy already exists
echo "Checking if policy $POLICY_NAME exists..."
aws iam get-policy --policy-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}" &> /dev/null

if [ $? -ne 0 ]; then
    echo "Policy $POLICY_NAME does not exist. Creating..."
    aws iam create-policy \
        --policy-name $POLICY_NAME \
        --policy-document file://$POLICY_DOCUMENT

    if [ $? -eq 0 ]; then
        echo "Policy $POLICY_NAME created successfully."
    else
        echo "Failed to create policy $POLICY_NAME. Exiting."
        exit 1
    fi
else
    echo "Policy $POLICY_NAME already exists."
fi

# Attach the policy to the role
echo "Attaching policy $POLICY_NAME to role $ROLE_NAME..."
aws iam attach-role-policy \
    --role-name $ROLE_NAME \
    --policy-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}"

if [ $? -eq 0 ]; then
    echo "Policy $POLICY_NAME attached to role $ROLE_NAME successfully."
else
    echo "Failed to attach policy $POLICY_NAME to role $ROLE_NAME. Exiting."
    exit 1
fi

# Clean up temporary files
rm -f $ASSUME_ROLE_POLICY_DOCUMENT $POLICY_DOCUMENT

echo "IAM role setup completed successfully."