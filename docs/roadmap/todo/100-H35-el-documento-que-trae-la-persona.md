---
id: H35
fase: "1"
---

# H35 · El documento que trae la persona, de cualquier tipo

- **Objetivo**: que lo que se dice de una póliza, un contrato o un convenio salga del documento que la persona trae, como lo que se dice de una sentencia (H23.1), y que la skill lo pida cuando pesa tanto como la ley.
- **Entrega**: el cotejo de `cita cotejar` generalizado a cualquier documento: huella, ficha y comprobación de que cada fragmento que se cita está en el documento; nada se guarda; las skills piden el documento cuando su contenido decide la respuesta y no lo suponen.
- **Alcance**: el PDF solo si el plan justifica la dependencia (H8 lo dejó fuera).
- **Controles**: golden de un fragmento que está y otro que no, en una sentencia y en una póliza; evals.
- [enmienda 2026-11-21, ADR 0059] El PDF entra en el alcance. El binario extrae y coteja el texto de un PDF que lo tiene, sin OCR y con la dependencia que elija el plan. Un PDF escaneado lo lee el modelo del agente, y lo que cite de él va marcado como no cotejado por el binario; el binario dice que no tiene capa de texto. H8 y H44 usan este cotejo.
