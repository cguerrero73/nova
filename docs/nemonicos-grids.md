# Nemónicos de Grids - Nova EAM

## Convenciones

### Estructura (6 caracteres)

```
X Y Y Y Y Y
│ │ └───────── Función (4 caracteres)
│ └─────────── Funcionalidad (1 caracter)
└───────────── Módulo (1 caracter)
```

### Módulo (1er caracter)

| Char | Módulo          |
| ---- | --------------- |
| `O`  | Objetos         |
| `S`  | Store (Almacén) |
| `W`  | Trabajo (Work)  |
| `B`  | Básico          |
| `P`  | Compras         |
| `J`  | Proyecto        |

### Funcionalidad (2do caracter)

| Char | Tipo        |
| ---- | ----------- |
| `M`  | Master data |
| `R`  | Reporte     |
| `N`  | Secundario  |

---

## Grids Definidos

| Nemónico | Grid        | Descripción         | Módulo  | Funcionalidad |
| -------- | ----------- | ------------------- | ------- | ------------- |
| `BMUSER` | Users       | Gestión de usuarios | Básico  | Master Data   |
| `SMPART` | Parts       | Partes/Repuestos    | Store   | Master Data   |
| `SMSTOR` | Stores      | Almacenes           | Store   | Master Data   |
| `OMOBJA` | Objects     | Objetos/Equipos     | Objetos | Master Data   |
| `WMJOBS` | Jobs/Events | Órdenes de trabajo  | Trabajo | Master Data   |
| `BCCODE` | Syscodes    | Códigos de sistema  | Básico  | Master Data   |
| `SMSTOC` | Stocks      | Existencias         | Store   | Master Data   |

---

## Reglas de Naming

1. Siempre 6 caracteres exactamente
2. Primeros 2 caracteres = prefijo (módulo + funcionalidad)
3. Últimos 4 caracteres = identificador de la función/pantalla
4. Usar mayúsculas
5. Evitar caracteres especiales

---

## Ejemplos de Referencia (R5/R-ONE)

| Nemónico Original | Descripción               |
| ----------------- | ------------------------- |
| `LVIRPART_MOB`    | Issue/Return Parts Mobile |
| `OCOBJC_RWO`      | Object Work Orders        |
| `WCJOBS_IPAD`     | Work Orders iPad          |
| `WCTASK_IPAD`     | Tasks iPad                |

---

_Última actualización: 2026-06-05_
