## Architecture Decisions

### Participant Status Management

Participant statuses (`confirmed`, `waitlist`, `dropped`, etc.) are **stored in the database** rather than calculated at runtime. This is a deliberate performance optimization for the game list view.

#### Why Store Status?

**Problem**: Calculating status at runtime would require:
- Fetching all participants for each game in the list (excluding inactive statuses like `dropped`)
- Ordering all participants by `joined_at` (first-come, first-served)
- Players at positions 1 through `max_participants` would be `confirmed`
- Players at positions beyond `max_participants` would be `waitlist`
- N+1 query problem: 20 games × 10 participants = 200+ records to fetch and process per list view

**Example**: For a game with `max_participants = 10`:
- Players who joined 1st-10th (ordered by `joined_at`) → `confirmed`
- Players who joined 11th+ → `waitlist`
- If player #5 drops, player #11 becomes #10 and gets promoted to `confirmed`

**Solution**: Store the status in the `participants.status` column and keep it synchronized using `reconcileParticipantStatuses`.

#### How We Keep Status Synchronized

The `reconcileParticipantStatuses` function is called immediately after any state-changing operation:

- **`JoinGame`** (line 820 in gamesservice.go): Reconciles after a user joins
- **`DropParticipantFromGame`** (line 935 in gamesservice.go): Reconciles after a user drops

This ensures:
1. Statuses are always accurate and never drift
2. The expensive computation happens on writes (join/drop), not on every list view
3. Game list queries are fast - just a simple JOIN to get `userParticipationStatus`
4. Users see correct status badges instantly

#### Performance Trade-off

- **Reads (game list)**: Fast - single query with JOIN
- **Writes (join/drop)**: Slightly slower - includes reconciliation, but writes are much less frequent than reads
- **Result**: Optimized for the most common operation (browsing games)

This is a classic **denormalization for read performance** pattern.

## Local Development
### Database

You can run a temporary postgres container to get you up and running without the headache of having postgres running using the following command:

```
docker run --name volley_db \
  -e POSTGRES_USER=volleyuser \
  -e POSTGRES_PASSWORD=volleyrocks123 \
  -e POSTGRES_DB=volley_dev \
  -p 5432:5432 \
  -d postgis/postgis:16-3.4
```

Postgres URLs follow the pattern:

```
postgresql://user:password@localhost:5432/mydb
```

From there, you set the `DATABASE_URL` environment variable like so (do NOT disable SSL mode in production):

```
export DATABASE_URL=postgresql://volleyuser:volleyrocks123@localhost:5432/volley_dev?sslmode=disable
```

Finally, run the app

```
export DATABASE_URL=postgresql://volleyuser:volleyrocks123@localhost:5432/volley_dev?sslmode=disable && \
go run ./...
```