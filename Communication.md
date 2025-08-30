# Comunicación Cliente - Servidor

- Todo tipo de mensaje enviado entre las entidades es mediante TCP, lo que garantiza la entrega de los mensajes.
- Los mensajes son en formato JSON.
- Ambas entidades consideran un mensaje como "completo" cuando reciben un salto de línea (`\n`).

## Formato mensaje de apuesta (de cliente a servidor)
```json
{
  "agency":"1",
  "birthdate":"1990-03-17"",
  "firstName":"Santiago Lionel",
  "id":"12345678",
  "lastName":"Lorca",
  "number":7574
}
```

## Formato mensaje de respuesta (de servidor a cliente)
```json
{
  "status":"OK", // o "FAIL"
}
```

> El formato de respuesta puede que sea modificado para puntos posteriores, ya que no es necesario para la consigna actual que tenga más información.
