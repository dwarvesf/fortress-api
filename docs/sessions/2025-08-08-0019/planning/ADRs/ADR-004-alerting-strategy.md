# ADR-004: Alerting Strategy and SLO Framework

**Status**: Proposed  
**Date**: 2025-08-08  
**Deciders**: Planning Team

## Context

The Fortress API serves critical business functions where downtime or degradation directly impacts:
- **Employee Operations**: Payroll processing, attendance tracking, performance reviews
- **Client Management**: Invoice generation, project delivery, billing processes  
- **Financial Systems**: Accounting automation, commission calculations, expense tracking
- **External Integrations**: Discord notifications, email delivery, file storage

Current alerting is reactive (manual monitoring) with no defined SLOs or automated incident detection. We need a proactive alerting strategy that balances early detection with alert fatigue prevention.

### Business Impact Analysis
- **High Impact**: Authentication failures, database outages, payment processing errors
- **Medium Impact**: External service degradation, slow query performance, elevated error rates
- **Low Impact**: Individual endpoint latency, non-critical feature failures

### Technical Constraints
- Existing Kubernetes cluster with Prometheus/Grafana/AlertManager
- Team operates in multiple timezones (some manual intervention delays)
- Mix of business and technical stakeholders need different alert contexts

## Decision

We will implement a **business-impact focused alerting strategy** using Service Level Objectives (SLOs) with multi-tier alerting based on severity and burn rate:

### 1. Service Level Objectives (SLOs)

#### Primary SLOs
```yaml
# 99.9% Availability SLO (43 minutes downtime/month)
availability_slo:
  target: 99.9%
  measurement: sum(rate(http_requests_total{status!~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
  error_budget: 0.1% # ~43 minutes/month
  
# 95th Percentile Latency SLO
latency_slo:
  target: 200ms  # 95% of requests under 200ms
  measurement: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
  
# Authentication Reliability SLO
auth_slo:
  target: 99.95% # Higher bar for auth
  measurement: sum(rate(auth_attempts_total{result="success"}[5m])) / sum(rate(auth_attempts_total[5m]))
  
# Database Health SLO
database_slo:
  target: 99.9%
  measurement: up{job="fortress-api-db"} # Connection health
```

#### Business-Specific SLOs
```yaml
# Payroll Processing (Critical Business Function)
payroll_slo:
  target: 99.99% # Zero tolerance during payroll periods
  measurement: sum(rate(business_operations_total{domain="payroll",result="success"}[5m])) / sum(rate(business_operations_total{domain="payroll"}[5m]))
  
# Invoice Generation (Revenue Critical)
invoice_slo:
  target: 99.9%
  measurement: sum(rate(business_operations_total{domain="invoice",result="success"}[5m])) / sum(rate(business_operations_total{domain="invoice"}[5m]))
```

### 2. Multi-Burn Rate Alerting Strategy

#### Fast Burn Alerts (Immediate Response)
```yaml
# Burn rate: 14.4x (consumes 1% error budget in 1 hour)
- alert: SLOAvailabilityFastBurn
  expr: |
    (
      1 - (
        sum(rate(http_requests_total{status!~"5.."}[5m])) / 
        sum(rate(http_requests_total[5m]))
      )
    ) > (14.4 * (1 - 0.999))
  for: 2m
  labels:
    severity: critical
    team: backend
    slo: availability
    burn_rate: fast
  annotations:
    summary: "High error rate is burning availability error budget fast"
    description: "{{ $value | humanizePercentage }} error rate over 5min (14.4x burn rate)"
    runbook_url: "https://wiki.company.com/slo-availability-fast-burn"
    slack_channel: "#alerts-critical"
```

#### Slow Burn Alerts (Planning Response)  
```yaml
# Burn rate: 6x (consumes 5% error budget in 1 day)
- alert: SLOAvailabilitySlowBurn
  expr: |
    (
      1 - (
        sum(rate(http_requests_total{status!~"5.."}[30m])) / 
        sum(rate(http_requests_total[30m]))
      )
    ) > (6 * (1 - 0.999))
  for: 15m
  labels:
    severity: warning
    team: backend
    slo: availability
    burn_rate: slow
  annotations:
    summary: "Sustained error rate is consuming availability error budget"
    description: "{{ $value | humanizePercentage }} error rate over 30min (6x burn rate)"
    runbook_url: "https://wiki.company.com/slo-availability-slow-burn"
    slack_channel: "#alerts-warning"
```

### 3. Comprehensive Alert Rules

#### Infrastructure Alerts (Critical)
```yaml
- alert: DatabaseConnectionsExhausted
  expr: |
    (gorm_dbstats_in_use / gorm_dbstats_max_open_connections) > 0.9
  for: 5m
  labels:
    severity: critical
    component: database
  annotations:
    summary: "Database connection pool near exhaustion"
    description: "{{ $value | humanizePercentage }} of database connections in use"
    
- alert: HighMemoryUsage
  expr: |
    (go_memstats_heap_alloc_bytes / go_memstats_heap_sys_bytes) > 0.85
  for: 10m
  labels:
    severity: warning
    component: runtime
  annotations:
    summary: "Application memory usage high"
    description: "{{ $value | humanizePercentage }} heap utilization"

- alert: GoroutineLeak
  expr: |
    go_goroutines > 1000
  for: 15m
  labels:
    severity: warning
    component: runtime
  annotations:
    summary: "Potential goroutine leak detected"
    description: "{{ $value }} goroutines running (threshold: 1000)"
```

#### Security Alerts (High Priority)
```yaml
- alert: AuthenticationFailureSpike
  expr: |
    rate(auth_attempts_total{result="failure"}[5m]) > 10
  for: 2m
  labels:
    severity: warning
    component: security
  annotations:
    summary: "High authentication failure rate detected"
    description: "{{ $value }} authentication failures per second"
    
- alert: SuspiciousActivityDetected
  expr: |
    rate(suspicious_activity_total[5m]) > 1
  for: 1m
  labels:
    severity: critical
    component: security
  annotations:
    summary: "Suspicious activity pattern detected"
    description: "{{ $value }} suspicious events per second"
    
- alert: PermissionViolationSpike
  expr: |
    rate(permission_checks_total{result="denied"}[5m]) > 5
  for: 3m
  labels:
    severity: warning
    component: security
  annotations:
    summary: "Elevated permission violations"
    description: "{{ $value }} permission denials per second"
```

#### Business Logic Alerts
```yaml
- alert: PayrollCalculationFailures
  expr: |
    rate(business_operations_total{domain="payroll",result="failure"}[5m]) > 0.1
  for: 1m
  labels:
    severity: critical
    component: payroll
    business_impact: high
  annotations:
    summary: "Payroll calculation failures detected"
    description: "{{ $value }} payroll failures per second"
    escalation: "@finance-team @engineering-lead"
    
- alert: InvoiceProcessingDegraded
  expr: |
    rate(business_operations_total{domain="invoice",result="failure"}[5m]) > 0.5
  for: 5m
  labels:
    severity: warning
    component: invoice
    business_impact: medium
  annotations:
    summary: "Invoice processing experiencing errors"
    description: "{{ $value }} invoice failures per second"
```

#### External Service Alerts
```yaml
- alert: ExternalServiceDown
  expr: |
    rate(external_api_calls_total{status!~"2.."}[5m]) / rate(external_api_calls_total[5m]) > 0.5
  for: 3m
  labels:
    severity: warning
    component: external
  annotations:
    summary: "External service {{ $labels.service }} experiencing issues"
    description: "{{ $value | humanizePercentage }} failure rate for {{ $labels.service }}"
    
- alert: EmailDeliveryFailures
  expr: |
    rate(external_api_calls_total{service="sendgrid",status!~"2.."}[5m]) > 1
  for: 5m
  labels:
    severity: warning
    component: email
    business_impact: medium
  annotations:
    summary: "Email delivery failures detected"
    description: "{{ $value }} email failures per second"
```

### 4. Alert Routing and Escalation

#### Severity-Based Routing
```yaml
# AlertManager configuration
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
  
  # Critical alerts -> immediate escalation
  - match:
      severity: critical
    receiver: 'critical-alerts'
    group_wait: 0s
    repeat_interval: 5m
    
  # Warning alerts -> team notifications  
  - match:
      severity: warning
    receiver: 'warning-alerts'
    repeat_interval: 30m
    
  # Business impact alerts -> stakeholder notifications
  - match:
      business_impact: high
    receiver: 'business-critical'
    group_wait: 0s

receivers:
- name: 'critical-alerts'
  slack_configs:
  - api_url: '{{ .SlackWebhookURL }}'
    channel: '#alerts-critical'
    title: 'CRITICAL: {{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
  pagerduty_configs:
  - routing_key: '{{ .PagerDutyKey }}'
    description: '{{ .GroupLabels.alertname }}'
    
- name: 'warning-alerts'
  slack_configs:
  - api_url: '{{ .SlackWebhookURL }}'
    channel: '#alerts-warning'
    
- name: 'business-critical'
  slack_configs:
  - api_url: '{{ .SlackWebhookURL }}'
    channel: '#business-alerts'
  email_configs:
  - to: 'finance-team@company.com, engineering-leads@company.com'
```

### 5. Alert Quality and Maintenance

#### Alert Review Process
- **Weekly Review**: Alert volume, false positive rate, resolution time
- **Monthly Calibration**: Threshold adjustments based on business patterns
- **Quarterly SLO Review**: Adjust targets based on business requirements

#### Alert Metrics (Monitoring the Monitors)
```yaml
# Alert effectiveness metrics
- alert: AlertManagerDown
  expr: up{job="alertmanager"} == 0
  for: 5m
  
- alert: HighAlertVolume
  expr: sum(rate(prometheus_notifications_total[1h])) > 50
  for: 10m
  annotations:
    summary: "Alert volume is high - potential alert storm or system issues"
    
- alert: LowSLOBurnRate
  expr: |
    (1 - slo_availability) < 0.0001  # Very low error rate
  for: 1d
  annotations:
    summary: "SLO targets may be too conservative - consider tightening"
```

### 6. Runbook Integration

#### Structured Runbooks
```markdown
# Runbook: SLO Availability Fast Burn

## Immediate Actions (< 5 minutes)
1. Check Grafana dashboard: [API Overview](http://grafana.company.com/d/api-overview)
2. Verify alert is not false positive
3. Identify failing endpoints from metrics
4. Check recent deployments: `kubectl rollout history deployment/fortress-api`

## Investigation Steps (5-15 minutes)  
1. Review error logs: `kubectl logs -f deployment/fortress-api --since=10m`
2. Check database connectivity: `kubectl exec -it postgres-pod -- pg_isready`
3. Review external service status pages
4. Check resource utilization: CPU, memory, connections

## Escalation Triggers
- Issue not resolved in 15 minutes -> Page engineering lead
- Database involved -> Page database team  
- Multiple services affected -> Page infrastructure team

## Post-Incident
- Update incident timeline in [Company Tool]
- Schedule post-mortem if >99.9% SLO violated
- Review and update alert thresholds if needed
```

### 7. Business Context Integration

#### Time-Aware Alerting
```yaml
# Payroll processing alerts - stricter during payroll periods
- alert: PayrollPeriodDegradation
  expr: |
    # More sensitive during last week of month
    (rate(business_operations_total{domain="payroll",result="failure"}[5m]) > 0.01)
    AND ON() (day_of_month() > 25)
  labels:
    severity: critical
    context: payroll_period
```

#### Business Hours Context
```yaml
# Different thresholds for business vs off hours  
- alert: HighLatencyBusinessHours
  expr: |
    histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 0.2
    AND ON() (hour() >= 9 AND hour() <= 17)  # 9 AM - 5 PM
  labels:
    severity: warning
    context: business_hours
    
- alert: HighLatencyOffHours  
  expr: |
    histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 0.5
    AND ON() (hour() < 9 OR hour() > 17)
  labels:
    severity: info
    context: off_hours
```

## Implementation Timeline

### Week 1: SLO Definition
- Define business-critical SLOs
- Implement basic error budget tracking
- Set up foundational alert rules

### Week 2: Alert Rules Implementation  
- Deploy comprehensive alert rules
- Configure AlertManager routing
- Test alert delivery mechanisms

### Week 3: Runbook Creation
- Create structured runbooks for each alert
- Train team on response procedures
- Set up escalation processes

### Week 4: Tuning and Optimization
- Monitor alert volume and effectiveness
- Adjust thresholds based on baseline data
- Document lessons learned

## Success Metrics

### Technical Metrics
- **Mean Time to Detection (MTTD)**: <5 minutes for critical issues
- **Alert Accuracy**: <5% false positive rate
- **Coverage**: 100% of critical business functions covered

### Operational Metrics
- **Mean Time to Acknowledgment**: <10 minutes during business hours
- **Mean Time to Resolution**: <30 minutes for service degradation
- **SLO Compliance**: >99% achievement of defined SLOs

### Business Metrics
- **Unplanned Downtime**: Zero incidents >15 minutes undetected
- **Customer Impact**: No customer-reported issues before internal detection
- **Business Continuity**: Zero payroll/invoice processing delays due to undetected issues

## Consequences

### Positive
- **Proactive Issue Detection**: Problems identified before business impact
- **Clear Escalation**: Structured response reduces confusion and delays
- **Business Alignment**: Alerts tied to actual business impact
- **Continuous Improvement**: SLO framework drives system reliability improvements

### Negative
- **Complexity**: More alert rules and processes to maintain
- **Learning Curve**: Team needs training on SLO concepts and tools
- **False Positives**: Initial period of threshold tuning
- **On-Call Overhead**: More structured on-call responsibilities

### Risks and Mitigations
- **Alert Fatigue**: Mitigated by business-impact focus and threshold tuning
- **Over-Engineering**: Balanced by starting simple and iterating
- **Missed Edge Cases**: Addressed through post-incident reviews and continuous refinement
- **Tool Complexity**: Training and documentation minimize learning curve

---

**Approved By**: Engineering & Operations Teams  
**Business Stakeholder**: Finance & HR Leadership  
**Implementation Owner**: SRE Team  
**Review Cadence**: Monthly SLO review, Weekly alert effectiveness review