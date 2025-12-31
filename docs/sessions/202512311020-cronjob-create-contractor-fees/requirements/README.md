# Requirements: Cronjob to Create Contractor Fees

## Overview

Create a cronjob endpoint under `/cronjobs` that automatically creates Contractor Fee entries from Task Order Logs with status "Approved".

## Functional Requirements

### FR1: Query Approved Task Order Logs
- Query Task Order Log database for entries with:
  - `Type` = "Order"
  - `Status` = "Approved"
- Extract relevant data: Contractor (via Deployment rollup), Date, Final Hours Worked, Proof of Works

### FR2: Find Matching Contractor Rate
- For each Task Order Log, find the corresponding Contractor Rate by:
  - Matching Contractor relation
  - Status = "Active"
  - Start Date <= Task Order Date <= End Date (or no End Date)

### FR3: Create Contractor Fee Entry
- Create a new Contractor Fee entry in Notion with:
  - `Task Order Log` relation → linked to the source Task Order Log
  - `Contractor Rate` relation → linked to the matching Contractor Rate
  - `Payment Status` = "New"
- Skip if Contractor Fee already exists for this Task Order Log

### FR4: Update Task Order Log Status
- After creating Contractor Fee, update Task Order Log status from "Approved" to "Completed"

## Non-Functional Requirements

### NFR1: Idempotency
- The cronjob must be idempotent - running multiple times should not create duplicate Contractor Fees
- Check if Contractor Fee already exists before creating

### NFR2: Logging
- DEBUG level logging for tracing
- Log each Task Order Log processed
- Log Contractor Fee creation success/failure

### NFR3: Error Handling
- Continue processing other entries if one fails
- Log errors but don't stop the entire job

## Data Sources

| Database | ID | Purpose |
|----------|-----|---------|
| Task Order Log | `2b964b29-b84c-801e-ab9e-000b0662b987` | Source of approved orders |
| Contractor Rates | `2c464b29-b84c-80cf-bef6-000b42bce15e` | Rate information |
| Contractor Fees | `2c264b29-b84c-8037-807c-000bf6d0792c` | Target for new entries |

## API Endpoint

- **Method**: POST
- **Path**: `/api/v1/cronjobs/contractor-fees`
- **Auth**: Cronjob authentication (API key or internal)
