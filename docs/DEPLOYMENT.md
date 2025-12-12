# Deployment Guide

## Prerequisites

- AWS Account with appropriate permissions
- Terraform >= 1.5
- kubectl >= 1.28
- Docker >= 24.0
- AWS CLI configured
- GitHub account for CI/CD

## Infrastructure Setup

### 1. Configure Terraform Backend

\`\`\`bash
# Create S3 bucket for Terraform state
aws s3api create-bucket \
  --bucket fintech-platform-terraform-state \
  --region us-east-1

# Create DynamoDB table for state locking
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
\`\`\`

### 2. Deploy Infrastructure

\`\`\`bash
cd infrastructure/terraform

# Initialize Terraform
terraform init

# Review plan
terraform plan -out=tfplan

# Apply infrastructure
terraform apply tfplan
\`\`\`

This will create:
- VPC with public/private subnets across 3 AZs
- EKS cluster with managed node groups
- RDS PostgreSQL database (Multi-AZ)
- ElastiCache Redis cluster
- Application Load Balancer
- S3 buckets
- IAM roles and policies
- Secrets Manager secrets

### 3. Configure kubectl

\`\`\`bash
aws eks update-kubeconfig \
  --name fintech-platform-production \
  --region us-east-1
\`\`\`

### 4. Deploy Kubernetes Resources

\`\`\`bash
# Create namespace
kubectl apply -f infrastructure/kubernetes/namespace.yaml

# Create ConfigMaps and Secrets
kubectl apply -f infrastructure/kubernetes/configmap.yaml
kubectl apply -f infrastructure/kubernetes/secrets.yaml

# Deploy applications
kubectl apply -f infrastructure/kubernetes/

# Verify deployments
kubectl get pods -n fintech-platform
\`\`\`

## Database Setup

### 1. Run Migrations

\`\`\`bash
# Get database endpoint
DB_ENDPOINT=$(terraform output -raw rds_endpoint)

# Run migrations
for file in database/migrations/*.sql; do
  psql -h $DB_ENDPOINT -U fintech -d fintech -f $file
done
\`\`\`

### 2. Seed Data (Optional)

\`\`\`bash
psql -h $DB_ENDPOINT -U fintech -d fintech -f database/seeds/initial_data.sql
\`\`\`

## CI/CD Setup

### 1. Configure GitHub Secrets

Add the following secrets to your GitHub repository:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_ACCOUNT_ID`

### 2. Enable GitHub Actions

The CI/CD pipeline will automatically:
- Run security scans on every push
- Run tests on pull requests
- Build and deploy on merge to main

### 3. Manual Deployment

\`\`\`bash
# Trigger deployment manually
gh workflow run ci-cd.yaml
\`\`\`

## Monitoring Setup

### 1. Access Grafana

\`\`\`bash
kubectl port-forward svc/grafana 3000:3000 -n monitoring
\`\`\`

Visit http://localhost:3000 (admin/admin)

### 2. Import Dashboards

Dashboards are automatically provisioned in `monitoring/grafana/dashboards/`

### 3. Configure Alerting

Update alert rules in `monitoring/prometheus/alerts/`

\`\`\`bash
# Reload Prometheus configuration
kubectl rollout restart deployment/prometheus -n monitoring
\`\`\`

## Scaling

### Manual Scaling

\`\`\`bash
# Scale a specific service
kubectl scale deployment credit-scoring --replicas=5 -n fintech-platform
\`\`\`

### Auto-scaling

HPA is configured by default:
- Min replicas: 3
- Max replicas: 10
- CPU threshold: 70%
- Memory threshold: 80%

## Blue-Green Deployment

\`\`\`bash
# Deploy new version
kubectl set image deployment/credit-scoring \
  credit-scoring=your-registry/credit-scoring:v2 \
  -n fintech-platform

# Monitor rollout
kubectl rollout status deployment/credit-scoring -n fintech-platform

# Rollback if needed
kubectl rollout undo deployment/credit-scoring -n fintech-platform
\`\`\`

## Security

### 1. Enable mTLS (Service Mesh)

\`\`\`bash
# Install Istio
istioctl install --set profile=production

# Enable mTLS
kubectl apply -f infrastructure/kubernetes/mtls-policy.yaml
\`\`\`

### 2. Rotate Secrets

\`\`\`bash
# Update secrets in AWS Secrets Manager
aws secretsmanager update-secret \
  --secret-id fintech-platform-production-db-credentials \
  --secret-string '{"password":"new-password"}'

# Restart pods to pick up new secrets
kubectl rollout restart deployment -n fintech-platform
\`\`\`

### 3. Certificate Management

Certificates are automatically managed by cert-manager using Let's Encrypt.

## Backup and Recovery

### Database Backups

Automated daily backups are configured for RDS with 7-day retention.

Manual backup:
\`\`\`bash
aws rds create-db-snapshot \
  --db-instance-identifier fintech-platform-production \
  --db-snapshot-identifier manual-backup-$(date +%Y%m%d)
\`\`\`

### Disaster Recovery

\`\`\`bash
# Restore from snapshot
terraform apply -var="restore_from_snapshot=snapshot-id"
\`\`\`

## Troubleshooting

### Pod Not Starting

\`\`\`bash
# Check pod status
kubectl describe pod <pod-name> -n fintech-platform

# Check logs
kubectl logs <pod-name> -n fintech-platform

# Check events
kubectl get events -n fintech-platform --sort-by='.lastTimestamp'
\`\`\`

### Database Connection Issues

\`\`\`bash
# Test connectivity from pod
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h <db-endpoint> -U fintech -d fintech
\`\`\`

### High Latency

1. Check Grafana dashboards
2. Review Jaeger traces
3. Check database slow query logs
4. Review Prometheus metrics

## Production Checklist

- [ ] All secrets rotated from default values
- [ ] SSL certificates configured
- [ ] Monitoring and alerting tested
- [ ] Backup strategy verified
- [ ] Disaster recovery plan documented
- [ ] Security scanning passing
- [ ] Load testing completed
- [ ] Documentation updated
- [ ] Team trained on deployment process
- [ ] Rollback procedure tested
