# ICY Token Transaction Analysis System Prompt

This system prompt will guide you step-by-step to run the necessary SQL queries to analyze ICY token transactions and create a comprehensive report.

## Step 1: Identifying Available Tables and Databases

First, let's check what databases are available. We need to use correct syntax for the SQL query tool:

```
{
  `query`: `SELECT datname FROM pg_database;`,
  `database`: `tono`
}
```

Since we can't run SHOW DATABASES directly, we need to explore the available tables in the database we're connected to:

```
{
  `query`: `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';`,
  `database`: `tono`
}
```

## Step 2: Inspect the Structure of the Community Token Transactions Table

Now that we've found the `community_token_transactions` table, let's examine its structure:

```
{
  `query`: `SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'community_token_transactions';`,
  `database`: `tono`
}
```

## Step 3: Identify the ICY Token ID

Let's find the most frequently used token ID, which should be the ICY token:

```
{
  `query`: `SELECT DISTINCT community_token_id, COUNT(*) as transaction_count FROM community_token_transactions GROUP BY community_token_id ORDER BY transaction_count DESC LIMIT 10;`,
  `database`: `tono`
}
```

Now we've confirmed that the ICY token ID is `9d25232e-add3-4bd8-b7c6-be6c14debc58`.

## Step 4: Check the BTC Deposit Date

Now we need to check the latest BTC deposit in the icy_swap database:

```
{
  `query`: `SELECT created_at FROM onchain_btc_transactions WHERE type = 'in' ORDER BY created_at DESC LIMIT 1;`,
  `database`: `icy_swap`
}
```

## Step 5: Get Transaction Type Summary

Now that we have the token ID and the start date, let's get the transaction type summary:

```
{
  `query`: `SELECT type, COUNT(*) as transaction_count, SUM(CAST(amount AS DECIMAL(36,18))) as total_amount FROM community_token_transactions WHERE community_token_id = '9d25232e-add3-4bd8-b7c6-be6c14debc58' AND created_at >= '2025-03-17T19:57:54.248Z' AND (type = 'airdrop' OR type = 'vault_transfer' OR (type = 'transfer' AND metadata::text LIKE '%\"reason\":\"distribute rewards\"%')) GROUP BY type ORDER BY total_amount DESC;`,
  `database`: `tono`
}
```

### Notes

- DO NOT include type 'withdraw'

## Step 6: Get Reward Category Breakdown

Let's get the breakdown of rewards by category:

```
{
  `query`: `SELECT category, COUNT(*) as transaction_count, SUM(CAST(amount AS DECIMAL(36,18))) as total_amount FROM community_token_transactions WHERE community_token_id = '9d25232e-add3-4bd8-b7c6-be6c14debc58' AND created_at >= '2025-03-17T19:57:54.248Z' AND type = 'transfer' AND metadata::text LIKE '%\"reason\":\"distribute rewards\"%' GROUP BY category ORDER BY total_amount DESC;`,
  `database`: `tono`
}
```

## Step 7: Get Vault Transfer Details

Let's get details about vault transfers:

```
{
  `query`: `SELECT sender_id, recipient_id, amount, description, created_at, metadata FROM community_token_transactions WHERE community_token_id = '9d25232e-add3-4bd8-b7c6-be6c14debc58' AND created_at >= '2025-03-17T19:57:54.248Z' AND type = 'vault_transfer' ORDER BY created_at;`,
  `database`: `tono`
}
```

## Step 8: Get Airdrop Details

Let's get details about airdrops:

```

{
  `query`: `SELECT sender_id, amount, description, created_at, metadata FROM community_token_transactions WHERE community_token_id = '9d25232e-add3-4bd8-b7c6-be6c14debc58' AND created_at >= '2025-03-17T19:57:54.248Z' AND type = 'airdrop' ORDER BY created_at;`,
  `database`: `tono`
}
```

## Step 9: Get BTC Deposit Information

Let's get details about the BTC deposit:

```
{
  `query`: `SELECT * FROM onchain_btc_transactions WHERE type = 'in' AND created_at >= '2025-03-17T19:57:54.248Z' ORDER BY created_at DESC LIMIT 5;`,
  `database`: `icy_swap`
}
```

## Step 10: Compile Final Report

After collecting all data, generate the final report with this format:

```markdown
# ICY Token Transaction Report (2025-03-17 - 2025-04-22)

## Total ICY minted: 652.80 ICY

### Transaction Type Summary
| Type | Count | Total Amount (ICY) |
|------|-------|-------------------|
| transfer | 526 | 580.80 |
| vault_transfer | 7 | 55.00 |
| airdrop | 20 | 17.00 |

### Reward Distribution By Category
| Category | Count | Total Amount (ICY) |
|----------|-------|-------------------|
|  | 158 | 474.00 |
| share a link | 350 | 105.00 |
| gain newbie role | 6 | 1.20 |
| s24-img-bounty | 12 | 0.60 |

### Notes
- BTC deposited on March 17, 2025 with a total of 1,697,000 Satoshi (0.01697 BTC).
```

This step-by-step guide shows you exactly how to extract and analyze ICY token transaction data from the databases. By following these steps, you can create a comprehensive report showing transaction summaries, reward distributions, and relevant BTC deposit information.