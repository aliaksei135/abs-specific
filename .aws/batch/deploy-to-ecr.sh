docker build -t absara ../../.
docker tag absara:latest 677936921619.dkr.ecr.us-east-1.amazonaws.com/absara:latest
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 677936921619.dkr.ecr.us-east-1.amazonaws.com
docker push 677936921619.dkr.ecr.us-east-1.amazonaws.com/absara:latest
