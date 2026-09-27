Guía de Especificaciones (Specs)

Bienvenido a la carpeta de especificaciones del proyecto. Este directorio sirve como la fuente de verdad para definir las funcionalidades, requerimientos y comportamiento esperado antes de que sean procesados e implementados por el agente.

🎯 Objetivo de esta Carpeta

En esta carpeta se almacenan todos los archivos de especificaciones (specs). Su propósito es servir como puente de comunicación entre la definición de producto/negocio y el agente de desarrollo.

Para garantizar el máximo desempeño y precisión del agente, es fundamental seguir las pautas descritas a continuación.

📋 Buenas Prácticas para Redactar Specs

1. Especificar Funciones Concretas

Evita descripciones ambiguas o demasiado generales. Define claramente:

El nombre y propósito de la función o módulo.

Las entradas (inputs) y salidas (outputs) esperadas.

Los casos de borde (edge cases) y el manejo de errores.

2. Claridad en el "Qué" y el "Cómo"

El agente genera mejores resultados cuando se le proporciona contexto claro:

Tener conocimiento claro de qué se quiere lograr: Define el objetivo de la funcionalidad sin vacilaciones.

Definir de qué forma se quiere hacer: Indica decisiones clave de diseño, flujos de datos o dependencias esperadas.

🛠 Uso de la Skill de Backend y Reglas del Proyecto

Para mantener la integridad de la aplicación durante la implementación:

Uso de la Skill de Backend: Al solicitar al agente que trabaje sobre una spec, asegúrate de activar la skill de backend. Esto garantiza que el agente siga el ciclo de Construcción → Optimización → Propuesta de Próximos Pasos.

Respeto a la Arquitectura: El agente se guiará estrictamente por la documentación ubicada en /docs/rules para no romper patrones arquitectónicos, convenciones de código ni estándares establecidos en el proyecto.

📝 Plantilla Sugerida para una Spec

# [Nombre de la Funcionalidad / Módulo]

## 1. Descripción General
[Explicación concisa de qué hace esta función y su objetivo]

## 2. Especificación Técnica
- **Entradas:** [Parámetros, payloads, etc.]
- **Salidas:** [Estructura de respuesta, tipos de datos]
- **Comportamiento:** [Paso a paso de la lógica interna]

## 3. Criterios de Aceptación
- [ ] Debe cumplir X requerimiento.
- [ ] Debe manejar el error Y de forma explícita.
