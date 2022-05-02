# go-ej1

## Como configurar
Toda la configuración se guarda en un archivo de formato yaml. Se puede ver un ejemplo de una configuración por defecto en resource/config/app.config


Esta configuración permite:
- Usar una lista fija para los lenguajes a buscar, definiendo ```usar_lista_fia: true``` y poniendo la lista como por ejemplo ```lista_lenguajes: [sle, python, c]```.
- Usar directamente la lista top20 de tiobe definiendo ```usar_lista_fia: false``` y definiendo las necesarias traducciones de tiobe a github en aliases (ver configuración por defecto para ejemplos).
- Definir el archivo donde se guarda el grafo
- Definir el archivo donde se guarda el resultado en texto

Además se pueden pasar los siguientes parametros en consola:
- ```-c <ARCHIVO CONFIGURACION>``` o ```--configfile <ARCHIVO CONFIGURACION>``` para el archivo de configuración. Por defecto se usa resource/config/app.config
- ```-l <LEVEL>``` o ```--loglevel <LEVEL>``` para el nivel de los logs mostrados. Por defecto se usa INFO. Las opciones son: ERROR, INFO, TRACE

## Como ejecutar
El repositorio ya incluye todas los modulos externos utilizado en la carpeta vendo. Por lo tanto se puede ejecutar directamente con ```go run .``` o compilar con ```go build``` y ejecutando el binario resultante.

Es importante mencionar que el grafo generado es en formato de una página web y requiere que por defecto sea configurado un navegador que permita la ejecución de código Javascript para la visualización. En caso contrario también existe la opción de abrir el archivo manualmente después de la ejecución con un programa adecuado (el nombre y dirección del archivo son configurables).
## Organización del repositorio
- No se hace push directo sobre main
- Para cada nueva función se crea un branch con nombre "feature/<funcion>"
- Para arreglar un bug se crea un branch con nombre "bugfix/<bug>"

  Cada mergerequest debe ser aprobado por al menos una persona (que no es el autor). Antes de un merge se actualiza CHANGELOG.md
