# Dashboard de <Proyecto>

## Tareas por estado

```dataview
TABLE length(rows) AS "Count"
FROM "03-tasks"
GROUP BY status
```

## Tareas por tipo

```dataview
TABLE
  length(rows) AS "Count",
  sum(rows.story_points) AS "Total SP"
FROM "03-tasks"
WHERE status != "done"
GROUP BY type
```

## Alta prioridad

```dataview
TABLE
  status AS "Status",
  type AS "Type",
  assignee AS "Assignee",
  due AS "Due"
FROM "03-tasks"
WHERE priority = "P0" AND status != "done"
SORT due ASC
```

## Creadas recientemente

```dataview
TABLE
  type AS "Type",
  status AS "Status",
  story_points AS "SP"
FROM "03-tasks"
SORT created DESC
LIMIT 10
```

## Bugs abiertos

```dataview
TABLE
  priority AS "Priority",
  service AS "Service",
  assignee AS "Assignee"
FROM "03-tasks"
WHERE type = "bug" AND status != "done"
SORT priority ASC
```
