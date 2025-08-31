# Comunicación Cliente - Servidor

- Todo tipo de mensaje enviado entre las entidades es mediante TCP, lo que garantiza la entrega de los mensajes.
- Los mensajes son en formato CSV.
- Ambas entidades consideran un mensaje como "completo" cuando reciben un salto de línea (`\n`).

## Formato mensaje de apuesta (de cliente a servidor)
```
1,Santiago Lionel,Lorca,12345678,1990-03-17,7574/n
```
Este formato CSV contiene los siguientes campos en orden:
1. agency (ID de la agencia)
2. firstName (Nombre)
3. lastName (Apellido)
4. document (DNI/ID)
5. birthdate (Fecha de nacimiento en formato YYYY-MM-DD)
6. number (Número de la apuesta)

## Formato mensaje de respuesta (de servidor a cliente)
```
OK/n
```

En caso de error:
```
FAIL,<error_message>/n
```

> El formato de respuesta puede que sea modificado para puntos posteriores, ya que no es necesario para la consigna actual que tenga más información.
