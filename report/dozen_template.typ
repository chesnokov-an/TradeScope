#let dozen_title_page(
  subjName: "",
  semNum: "",
  theme: "",
  student: "",
  studentInitials: "",
  group: "",
  tutor: "",
  tutorInitials: "",
  year: "",
) = {
  set page(paper: "a4", margin: (left: 3cm, right: 1.5cm, top: 2cm, bottom: 2cm))
  set text(font: ("Liberation Serif"), size: 12pt, lang: "ru")
  set par(leading: 0.8em)
  
  align(center)[
    #text(size: 11pt)[МИНИСТЕРСТВО НАУКИ И ВЫСШЕГО ОБРАЗОВАНИЯ РОССИЙСКОЙ ФЕДЕРАЦИИ]
    #v(0.3em)
    #text(size: 8pt)[ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ АВТОНОМНОЕ ОБРАЗОВАТЕЛЬНОЕ УЧРЕЖДЕНИЕ ВЫСШЕГО ОБРАЗОВАНИЯ]
  
    #v(0.5em)
    #text(size: 12pt, weight: "bold")[«Национальный исследовательский ядерный университет «МИФИ»]
    #text(size: 12pt, weight: "bold")[(НИЯУ МИФИ)]
      #line(length: 470pt, stroke: 0.5pt)
    #v(1.2em)
    #text(size: 12pt)[ИНСТИТУТ ИНТЕЛЛЕКТУАЛЬНЫХ КИБЕРНЕТИЧЕСКИХ СИСТЕМ]
    #v(0.5em)
    #text(size: 12pt)[КАФЕДРА «КОМПЬЮТЕРНЫЕ СИСТЕМЫ И ТЕХНОЛОГИИ» (№12)]
  ]
  
  
  align(center)[
    #v(1cm)
    #text(size: 16pt, weight: "bold")[ОТЧЕТ]
    
    #text(size: 14pt, weight: "bold")[КУРСОВОЙ ПРОЕКТ]

    #text(size: 14pt, weight: "bold")[(#subjName)]
    
    #text(size: 12pt, weight: "bold")[Семестр #semNum]
  ]
  
  
  v(1cm)
  
  table(
    columns: (3.5cm, 9cm, 2.5cm, 1.8cm),
    stroke: none,
    inset: (x: 6pt, y: 8pt),
    align: (left, left, left, left),
    table.cell(colspan: 4)[*Тема:* "#theme"],
    table.hline(stroke: 0.5pt),
    [Студент], [#student], [Группа], [#group],
    [], [#v(-0.6cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_], [], [],
    [], [#align(center)[#v(-0.6cm)ФИО]], [], [],
    [#v(-0.3cm)Руководитель],
      table.cell(colspan: 3)[#v(-0.3cm)#tutor],
      [], [#v(-0.6cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_], [],
    [],
      table.cell(colspan: 3)[#align(center)[#v(-0.6cm)ФИО, должность, звание, степень#h(-5cm)]],
  )
  
  table(
    columns: (5.5cm, 3.5cm, 2cm, 6cm),
    stroke: none,
    inset: (x: 6pt, y: 8pt),
    align: (left, center, left, right),
    [#v(-0.6cm)Студент],          [#v(-0.6cm)],            [#v(-0.6cm)], [#v(-0.6cm)#studentInitials],
    [],[#v(-0.6cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_#h(-5.4cm)], [], [],
    [],                 [#v(-0.6cm)подпись#h(-2.2cm)],     [], [#v(-0.6cm)ФИО#h(0.7cm)],
    
    [Руководитель],     [],            [], [#tutorInitials],
    [],[#v(-0.6cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_#h(-5.4cm)], [], [],
    [],                 [#v(-0.6cm)подпись#h(-2.2cm)],     [], [#v(-0.6cm)ФИО#h(0.7cm)],
    [#h(1.15cm)*Оценка*],         [#v(0.1cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_#h(2cm)],            [], [],
    [*Баллы, ECTS*],    [#v(0.1cm)\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_#h(2cm)],            [], [],
  )
  table(
    columns: (12.0cm, 5cm),
    stroke: none,
    inset: (x: 6pt, y: 0pt),
    align: (left, center, left, right),
    [], [#v(-0.1cm) «\_\_\_\_»\_\_\_\_\_\_\_\_\_\_#year г.]
  )
  
  v(2.5cm) // если москва уехала, то измени этот параметр
  align(center)[
    #text(size: 12pt, weight: "bold")[Москва #year г.]
  ]
  pagebreak()
}

#let report(
  subjName: "",
  semNum: "",
  theme: "",
  student: "",
  studentInitials: "",
  group: "",
  tutor: "",
  tutorInitials: "",
  year: "",
  content
) = {
  set page(
    margin: (
      top: 2cm, 
      bottom: 2cm, 
      left: 2cm, 
      right: 2cm,
    ),
  )
  set text(
    size: 12pt,
    lang: "ru",
  )
  set heading(
    numbering: "1.1",
    hanging-indent: 1pt,
  )
  show heading: it => {
    it
    set text(1pt) // these two lines are kostil' 
    linebreak()
  }
  set par(
    first-line-indent: 1.5cm,
    leading: 1em,
    spacing: 1em,
    justify: true,
  )
  set enum(
    indent: 1.5em,
  )
  dozen_title_page(
  subjName: subjName,
  semNum: semNum,
  theme: theme,
  student: student,
  studentInitials: studentInitials,
  group: group,
  tutor: tutor,
  tutorInitials: tutorInitials,
  year: year,
  )
  set page(
    numbering: "1",
  )
  counter(page).update(2)

  content
}

#let dozen_image(
  file, 
  height: auto, 
  width: auto, 
  caption: [],
) = {
  set text(12pt)
  set figure.caption(
    separator: [ -- ],
  )
  figure(
    supplement: "Рисунок", 
    numbering: "1",
    caption: caption,
    image(
      file, 
      height: height, 
      width: width,
    )
  )
}

#let dozen_table(
  body,
  caption: [],
  lbl: none,
) = {
  set figure.caption(
    separator: [ -- ], 
    position: top,
  )
  let fig = figure(
    supplement: "Таблица", 
    numbering: "1",
    caption: caption,
    body
  )
  if lbl == none [ #fig ] else [ #fig #label(lbl) ]
}

#let dozen_code(
  caption: [], 
  body
) = {
  set raw(
    tab-size: 4,
  )
  set text(size: 8pt)
  show raw.where(block: true) : it => block(
      fill: luma(240),
      inset: 10pt,
      outset: (x: 1pt, y: 1pt),
      radius: 5pt,
      it
  )
  set figure.caption(
    separator: [ -- ],
    position: top,
  )
  show figure: set block(breakable: true)
  figure(
    kind: raw,
    supplement: "Листинг", 
    caption: caption,
    table(
      columns: (1fr),
      stroke: none,
      align: left,
      [#body]
    )
  )
}