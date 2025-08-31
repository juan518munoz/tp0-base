# Comunicación Cliente - Servidor

- Todo tipo de mensaje enviado entre las entidades es mediante TCP, lo que garantiza la entrega de los mensajes.
- Los mensajes son en formato CSV.
- Los mensajes enviados por el cliente al servidor se consideran "completos" cuando se recibe un carácter nulo (`\0`).
- Los mensajes enviados por el servidor al cliente se consideran "completos" cuando se recibe un salto de línea (`\n`).

## Formato mensaje de apuesta (de cliente a servidor)

```
1,3
Santiago Lionel,Lorca,12345678,1990-03-17,7574
Ana María,Gómez,23456789,1985-05-22,1234
Carlos,Fernández,34567890,1992-11-09,5678\0
```
El formato comienza con el ID de la agencia y la cantidad de apuestas, seguido por cada apuesta en líneas separadas.

## Formato mensaje de respuesta (de servidor a cliente)
```
OK/n
```

En caso de error:
```
FAIL,<error_message>/n
```

> El formato de respuesta puede que sea modificado para puntos posteriores, ya que no es necesario para la consigna actual que tenga más información.
