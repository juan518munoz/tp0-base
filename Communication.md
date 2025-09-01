# Comunicación Cliente - Servidor

- Todo tipo de mensaje enviado entre las entidades es mediante TCP, lo que garantiza la entrega de los mensajes.
- Los mensajes son en formato CSV.
- Los mensajes enviados por el cliente al servidor se consideran "completos" cuando se recibe un carácter nulo (`\0`).
- Los mensajes enviados por el servidor al cliente se consideran "completos" cuando se recibe un salto de línea (`\n`).

## Tipos de mensajes
Existen tres tipos de mensajes que el cliente le puede enviar al servidor.

### Envio de apuestas
```csv
BETS,<agency_id>,<bet_count>
<FirstName>,<LastName>,<Document>,<Birthdate>,<Number>
<FirstName>,<LastName>,<Document>,<Birthdate>,<Number>
...
<FirstName>,<LastName>,<Document>,<Birthdate>,<Number>
```

**El servidor recibe este mensaje, en caso de que todas las apuestas enviadas sean validas y las almacena correctamente, responde con:**
```csv
OK
```

**Caso contrario, el error será notificado con el mensaje:**
```csv
FAIL,<error_msg>
```

### Envio notificación finalización de apuestas
```csv
FINISHED,<agency_id>
```

El servidor responde al cliente de la misma forma que con el envio de respuestas.

### Consulta resultados
```csv
RESULTS,<agency_id>
```

La respuesta del servidor depende de si el sorteo ya ha sido realizado.

**Si ya fue realizado:**
```csv
OK,<won_bets_count>
```

**Si aun hay agencias pendientes a finalizar:**
```csv
FAIL,NOT_READY
```

**Si la agencia solicitada no es valida:**
```csv
FAIL,INVALID_INPUT
```
