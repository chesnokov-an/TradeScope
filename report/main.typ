#import "dozen_template.typ": *
#import "generated/results.typ": *

#show: report.with(
  subjName: "ЛОГИЧЕСКИЕ ОСНОВЫ СИСТЕМ СПЕЦИАЛЬНОГО НАЗНАЧЕНИЯ С ЭЛЕМЕНТАМИ СИСТЕМНОГО АНАЛИЗА",
  semNum: "5",
  theme: "Анализ специальных торговых стратегий",
  student: "Чесноков Александр Николаевич",
  studentInitials: "Чесноков А. Н.",
  group: "С24-501",
  tutor: "Кулик Сергей Дмитриевич, профессор кафедры №12 НИЯУ МИФИ",
  tutorInitials: "Кулик С. Д.",
  year: "2026",
)

#set text(font: "Liberation Serif", size: 12pt)
#set math.equation(numbering: "(1)")

#align(center)[#text(size: 16pt, weight: "bold")[ОГЛАВЛЕНИЕ]]
#outline(title: [], depth: 3)
#pagebreak()

#include "chapters/01_introduction.typ"

#pagebreak()
#include "chapters/02_chapter1.typ"

#pagebreak()
#include "chapters/03_chapter2.typ"

#pagebreak()
#include "chapters/04_conclusion.typ"

#pagebreak()
#heading(numbering: none, outlined: true)[СПИСОК ИСПОЛЬЗОВАННЫХ ИСТОЧНИКОВ]

#show bibliography: it => {
  show grid.cell.where(x: 0): cell => {
    show regex("\[(\d+)\]"): s => s.text.slice(1, -1) + "."
    cell
  }
  it
}

#bibliography("bibliography.yml", style: "ieee", title: none)

#pagebreak()
#heading(numbering: none, outlined: true)[ПРИЛОЖЕНИЕ. ОСНОВНЫЕ ФРАГМЕНТЫ ИСХОДНОГО КОДА]
#include "chapters/05_appendix.typ"
