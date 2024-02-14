// Extrato e Saldo unificados

import "join"

transacoes = from(bucket: "rinha")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "transacoes")
  |> filter(fn: (r) => r["cliente"] == "1")
  |> pivot(rowKey:["_time","cliente"], columnKey: ["_field"], valueColumn: "_value")

creditos = transacoes
  |> filter(fn: (r) => r["tipo"] == "c")
  |> group(columns: ["cliente","tipo"])
  |> drop(columns: ["_start","_stop"])
  |> sort(columns: ["_time"], desc: true)

debitos = transacoes
  |> filter(fn: (r) => r["tipo"] == "d")
  |> group(columns: ["cliente","tipo"])
  |> drop(columns: ["_start","_stop"])
  |> sort(columns: ["_time"], desc: true)

creditos_sum =
  creditos
  |> sum(column: "valor")
  |> map(fn: (r) => ({
    r with
    descricao: "TOTAL CREDITOS",
    _time: now(),
    _measurement: "transacoes",
  }))

debitos_sum =
  debitos
  |> sum(column: "valor")
  |> map(fn: (r) => ({
    r with
    descricao: "TOTAL DEBITOS",
    _time: now(),
    _measurement: "transacoes",
  }))

consolidado = union(tables: [creditos,debitos,creditos_sum,debitos_sum])
|> group(columns: ["cliente"])
|> yield(name: "consolidado")

cliente = from(bucket: "rinha")
|> range(start: v.timeRangeStart, stop: v.timeRangeStop)
|> filter(fn: (r) => r["_measurement"] == "clientes")
|> filter(fn: (r) => r["cliente"] == "1")
|> filter(fn: (r) => r["_field"] == "limite")
|> keep(columns: ["cliente","_value"])

join.inner(
    left: creditos_sum
    |> group(columns: ["cliente"])
    |> drop(columns: ["tipo"]),
    right: debitos_sum
    |> group(columns: ["cliente"])
    |> drop(columns: ["tipo"]),
    on: (l,r) => l.cliente == r.cliente,
    as: (l,r) => ({
      l with
      total: int(v: l.valor) - int(v: r.valor),
      descricao: "SALDO",
      _time: now(),
    }),
    )
    |> drop(columns: ["valor"])
|> join.inner(
    right: cliente, 
    on: (l,r) => l.cliente == r.cliente,
    as: (l,r) => ({
        l with
        limite: r._value,
    })
)
|> yield(name: "saldo")
```

// Transações


transacoes = from(bucket: "rinha")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "transacoes")
  |> filter(fn: (r) => r["cliente"] == "1")
  |> pivot(rowKey:["_time","cliente"], columnKey: ["_field"], valueColumn: "_value")
  |> group(columns: ["cliente","tipo"])
  |> drop(columns: ["_start","_stop"])
  |> sort(columns: ["_time"], desc: true)


// Saldo

import "join"

transacoes = from(bucket: "rinha")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "transacoes")
  |> filter(fn: (r) => r["cliente"] == "1")
  |> pivot(rowKey:["_time","cliente"], columnKey: ["_field"], valueColumn: "_value")
  |> group(columns: ["cliente","tipo"])
  |> drop(columns: ["_start","_stop"])

creditos = transacoes 
|> filter(fn: (r) => r["tipo"] == "c")
|> sum(column: "valor")
|> drop(columns: ["tipo"])

debitos = transacoes 
|> filter(fn: (r) => r["tipo"] == "d")
|> sum(column: "valor")
|> drop(columns: ["tipo"])

join.inner(
    left: creditos,
    right: debitos,
    on: (l,r) => l.cliente == r.cliente,
    as: (l,r) => ({
        cliente: l.cliente,
        total: l.valor - r.valor,
    })
)