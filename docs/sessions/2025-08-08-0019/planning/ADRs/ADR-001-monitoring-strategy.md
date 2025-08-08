# ADR-001: Monitoring Strategy for Fortress API

**Status**: Proposed  
**Date**: 2025-08-08  
**Deciders**: Planning Team

## Context

The Fortress API is a production Go web API serving multiple critical business functions including HR management, project tracking, payroll processing, and client billing. Currently, the system has minimal observability beyond basic logging and a simple health check endpoint.

### Current State
- **Framework**: Gin web framework with layered architecture
- **Database**: PostgreSQL with GORM ORM
- **Infrastructure**: Kubernetes deployment with existing Prometheus/Grafana/Loki stack
- **Authentication**: JWT + API Key based with permission middleware
- **Monitoring**: Basic `/healthz` endpoint only

### Requirements
- Production-ready monitoring for performance, security, and business insights
- Integration with existing Kubernetes Prometheus/Grafana/Loki stack
- Minimal complexity, high value approach
- Open-source solutions only
- Alignment with existing architectural patterns

## Decision

We will implement a **layered monitoring strategy** using the Prometheus/Grafana/Loki stack with the following approach:

### 1. Instrumentation Strategy
- **HTTP Layer**: Gin middleware for request metrics (latency, throughput, errors)
- **Database Layer**: GORM plugin for connection pool and query performance  
- **Business Layer**: Custom metrics for authentication, authorization, and business operations
- **Application Layer**: Go runtime metrics (memory, GC, goroutines)
- **Security Layer**: Authentication events, permission failures, rate limiting violations

### 2. Technology Stack
- **Metrics Collection**: Prometheus Go client library (`github.com/prometheus/client_golang`)
- **HTTP Instrumentation**: Custom Gin middleware based on proven patterns
- **Database Monitoring**: Official GORM Prometheus plugin  
- **Structured Logging**: JSON-formatted logs with contextual fields for Loki
- **Service Discovery**: Kubernetes ServiceMonitor CRDs for automatic discovery

### 3. Metrics Categories Priority
**Tier 1 (Essential)**:
- HTTP request rate, latency, errors by endpoint
- Database connection pool utilization
- Go runtime metrics (heap, GC, goroutines)
- Authentication success/failure rates

**Tier 2 (Important)**:
- Business transaction metrics (invoice processing, payroll calculations)
- External API call success rates (Discord, SendGrid, GitHub)
- Security events (permission failures, suspicious patterns)

**Tier 3 (Valuable)**:
- Feature usage analytics
- User session patterns
- Performance optimization insights

## Rationale

### Why Prometheus + Grafana + Loki
1. **Existing Infrastructure**: Stack already deployed in Kubernetes
2. **Industry Standard**: Proven pattern for Go applications  
3. **Native Kubernetes Integration**: ServiceMonitor CRDs for auto-discovery
4. **Operational Simplicity**: Team already familiar with tools
5. **Cost Effective**: Open-source with no licensing costs

### Why Custom Gin Middleware
1. **Architecture Alignment**: Fits naturally into existing middleware chain
2. **Granular Control**: Can instrument specific business logic
3. **Performance**: Minimal overhead (<1ms per request)
4. **Maintainability**: Follows established patterns in codebase

### Why GORM Plugin  
1. **Official Support**: Maintained by GORM team
2. **Zero Configuration**: Automatic connection pool metrics
3. **Integration**: Seamless with existing database setup
4. **Reliability**: Production-tested across many applications

## Implementation Approach

### Phase 1: Foundation (Week 1)
- Add Prometheus client libraries
- Implement basic HTTP metrics middleware
- Expose `/metrics` endpoint
- Deploy basic Grafana dashboard

### Phase 2: Database & Business Metrics (Week 2)  
- Add GORM Prometheus plugin
- Implement authentication/authorization metrics
- Configure Kubernetes ServiceMonitor
- Set up alerting rules

### Phase 3: Structured Logging (Week 3)
- Implement JSON structured logging
- Configure Loki log aggregation  
- Add security event logging
- Create comprehensive dashboards

### Phase 4: Production Readiness (Week 4)
- Performance optimization
- Alert tuning and runbook creation
- Team training and documentation
- Production rollout

## Consequences

### Positive
- **Comprehensive Observability**: Full visibility into application performance and health
- **Proactive Issue Detection**: Mean Time to Detection (MTTD) < 5 minutes
- **Business Insights**: Understanding of feature usage and user patterns  
- **Security Monitoring**: Detection of authentication anomalies and attacks
- **Operational Excellence**: Data-driven optimization and capacity planning

### Negative  
- **Initial Development Effort**: ~3-4 weeks of focused development
- **Runtime Overhead**: ~1-2% CPU impact for metrics collection
- **Storage Requirements**: Additional storage for metrics and logs
- **Complexity**: New monitoring components to maintain and troubleshoot

### Risks & Mitigations
- **High Cardinality**: Controlled through endpoint normalization and label limits
- **Performance Impact**: Load testing and gradual rollout to validate
- **Alert Fatigue**: Careful threshold tuning and business-impact focus
- **Team Learning Curve**: Training sessions and comprehensive documentation

## Monitoring Goals

### Service Level Objectives (SLOs)
- **Availability**: 99.9% of requests succeed (non-5xx status codes)
- **Latency**: 95% of requests complete within 200ms
- **Error Rate**: <0.1% monthly error budget
- **MTTD**: <5 minutes for critical issues
- **MTTR**: <30 minutes for service degradation

### Success Metrics
- Zero unplanned downtime incidents go undetected  
- 90% reduction in time to identify performance bottlenecks
- Proactive identification of security threats
- Business insights drive at least 2 feature optimization decisions per quarter

## Next Steps

1. **Technical Specification Creation**: Detailed implementation specs for each component
2. **Proof of Concept**: Basic HTTP metrics in development environment
3. **Team Review**: Technical review and approval from engineering leads  
4. **Resource Planning**: Sprint planning and developer assignment
5. **Staging Deployment**: Validation in staging environment before production

---

**Decision Makers**: Engineering Team  
**Review Date**: 2025-09-08 (1 month post-implementation)