## Objectives
* RemoveCoins service may be doing more work than we'd like to retrieve coins randomly
* To scale, with concurrent Pyggpot processes

## Remove coin by kind specification

Add coins (Coinis) field to RemoveCoinRequest to specify the coins to be removed
```
message RemoveCoinsRequest {
    int32 pot_id = 1 [(validator.field) = {int_gt: 0}]; // required
    int32 count = 2 [(validator.field) = {int_gt: 0}];
    repeated Coins coins = 3;
}
```

### Validation Rule
* If `coins` field is not empty, then `count` field can be zero
* If `coins` field is empty, then `count` field value must be greater than 0
* Validation error when `coins` fields is empty AND `count` field less than or equal to 0

### Business Rule
* Remove no more than available coins of the respective kind
* Update coins records only on the changed kind
* Report count of the removed coins in response

### Performance Optimization
* models.CoinByID call can retrieve coins count concurrently
* Coins can be updated concurrently

### Coverage Test
* Above 90% code coverage
---

## To scale, with concurrent Pyggpot processes
Concurrent Pyggpot process means having more than one independent service instances running in a cluster.
In another world, multiple conflicting requests can be serviced by different service instances at the same time.

### Assumption
* Assume it is not common to have too many concurrent requests operating on the **same** Pyggpot.

### Business Rule
* Since Pyggpot is a unique container, its contents must be manipulated exclusively per request.
* Process should obtain lock over a Pyggpot before applying change content operations.
* Process should release lock when operation is over.
* Lock expiration policy should be attached to the lock to prevent deadlock by crashed process.
* Lock can be a UUID

### Table schema design
```sql
CREATE TABLE IF NOT EXISTS pot (
  id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  pot_name text NOT NULL UNIQUE,
  max_coins integer NOT NULL,
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  lock UUID NULL,
  lock_expire_time DATETIME NULL 
);

CREATE INDEX pot_lock ON pot (id, lock);
```

### Data Operation
#### Set Lock
```sql
UPDATE pot
set lock = $2, lock_expire_time = $3
WHERE id = $1 and (
    lock is NULL or lock_expire_time <= NOW())
```

#### Release Lock
```sql
UPDATE pot
set lock = NULL, lock_expire_time = NULL
WHERE id = $1 and lock = $2
```

#### Select Query
Select query does not require locking

#### Execute Data change Query
```go
func exec() error {
    tx := s.DB.Begin()
    committed := false
    defer func () {
    if !committed {
        _ = tx.Rollback()
    }
    }()

    // Set Lock

    // Execute data change query

    // Execute separate change query

    // Release Lock

    err := tx.Commit()
    if err != nil {
        return err
    }
    committed = true

    return nil
}
```

Data change query example
```sql
UPDATE coin
set coin.coin_count = $3
FROM coin INNER JOIN pot on coin.pot_id = pot.id
WHERE coin.id = $1 AND pot.lock = $2
```

#### Lock conflict resolution
In case of set Lock conflict, use timer to wait for n millisecond and then retry for r time.
Be aware timer may cause memory leak 
https://www.fatalerrors.org/a/use-with-caution-time.after-can-cause-memory-leak-golang.html
