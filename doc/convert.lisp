
(in-package :cl-user)

(named-readtables:in-readtable :buddens-readtable-a)

(proclaim '(optimize debug))
(declaim (optimize debug))

;; дя = данные-ячейки
(eval-when (:compile-toplevel :execute)
  (defstruct Дя
    Текст
    Url
    Комментарий)

  (defmethod make-load-form ((self Дя) &optional environment)
    (make-load-form-saving-slots self :environment environment))

  (defstruct Пользователь
    Id
    Nickname
    Registrationemail)

  (defmethod make-load-form ((self Пользователь) &optional environment)
    (make-load-form-saving-slots self :environment environment))

  (defstruct Диалект
    Id ; номер колонки
    Slug
    Commentary
    Ownerid)

  (defmethod make-load-form ((self Пользователь) &optional environment)
    (make-load-form-saving-slots self :environment environment))

  (defstruct Lws
    Languageid
    Word
    Senseid
    Commentary)

  (defmethod make-load-form ((self Lws) &optional environment)
    (make-load-form-saving-slots self :environment environment))

)

(defparameter *xml* 
  (with-open-file (s "c:/promo.yar/google-to-semdict/Англо-Русский\ словарь\ терминов\ и\ слов\ для\ включения\ в\ программы\ -\ 2021-03-29-правленый.fods")
(xmls:parse s)))


(defun имя-тега (тег)
  (caar тег))

(defun список-атрибутов-тега (тег имя-тега)
  "Имя-тега задаётся для проверки, что это действительно правильно переданный
тег. Если несколвозможно более одного вида тега, надо вернуть nil"
  (when имя-тега
    (assert (string= (имя-тега тег) имя-тега)))
  (second тег))

(defun значение-атрибута-тега (тег имя-тега имя-атрибута)
  "Имя-тега задаётся для проверки, что это действительно правильно переданный
тег. Из-за того XML игнорирует пр-ва имён, атрибуты могут дублироваться - в этом случае возвращается первый попавшийся"
  (let ((атрибуты (список-атрибутов-тега тег имя-тега)))
    (second (assoc имя-атрибута атрибуты
                   :test 'string=))))

;;; инспектируем *xml* и достаём оттуда 

(defparameter *body* (nth 9 *xml*))

(defparameter *spreadsheet* (nth 2 *body*))

(defparameter *лист-словарь* (nth 4 *spreadsheet*))

;;; Колонки и строки вынимаем руками

(defparameter *колонки*
  (subseq *лист-словарь* 3 14))

(defparameter *строки*
  (subseq *лист-словарь* 14 139))

(defparameter *строка-имён-языков* 
  (nth 0 *строки*))

(defparameter *индекс-строки-с-первым-смыслом* 1) 

(defun формула-только-url (ячейка)
  "Возвращает два значения - url, отформатированный как html, и текст"
  (let* ((список-где-формула (fourth (second ячейка)))
         (возможно-слово-формула (car список-где-формула))
         (формула 
          (and 
           (equal возможно-слово-формула "formula")
           (second список-где-формула))))
    (and формула (формулу-в-html формула))))

(defun формулу-в-html (формула)
  (perga-implementation:perga
   (let seq (split-sequence:split-sequence-if
             (lambda (ch) (find ch "\"")) формула))
   (assert
    (equal (first seq) "of:=HYPERLINK("))
   (let url (second seq))
   (assert
    (equal (third seq) ";"))
   (let text (fourth seq))
   (assert
    (equal (fifth seq) ")"))
   (values 
    (format nil "<a href=\"~A\">~A</a>" url text)
    text)))     


(defun очисть-абзац-от-формата (абзац &key комментарий)
  (perga-implementation:perga
   (cond 
    (комментарий
     (assert (equal (caar абзац) "annotation")))
    (t
     (assert (equal (caar абзац) "p"))))
   (let ч абзац)
   (loop
     (cond
      ((stringp ч)
       (return-from очисть-абзац-от-формата ч))
      ((stringp (car ч))
       (return-from очисть-абзац-от-формата (car ч)))
      ((equalp (caar ч) "p")
       (setq ч (third ч)))
      ((and комментарий
            (consp (caar ч))
            (equal (caaar ч) "p"))
       (setq ч (car ч)))
      ((equal (caar ч) "a")
       (format t "~&<a>: ~S~%" ч)
       (setq ч (third ч)))
      ((and (consp (car ч))
            (consp (caar ч))
            (equal (caaar ч) "style-name"))
       (pop ч))
      ((and (consp (car ч))
            (equal (caar ч) "span"))
       (setq ч (cdr ч)))
      ((and (consp (car ч))
            (consp (caar ч))
            (equal (caaar ч) "span"))
       (setq ч (cdar ч)))
      ((and комментарий
            (equalp (caar ч) "annotation"))
       (pop ч))
      ((and комментарий
            (consp (caar ч))
            (equalp (caaar ч) "caption-point-y"))
       (pop ч))
      ((and комментарий
            (consp (caar ч))
            (equalp (caaar ч) "date"))
       (pop ч))
      (t (error "Не знаю, что делать с ~S" ч))))))

(print (очисть-абзац-от-формата '(("p" . "неважно2") (("style-name" "P1"))
    (("span" . "неважно2") (("style-name" "T3")) "Церковнославянское \"как\"."))))
    
   
(defun первый-комментарий (ячейка)
  "Если ячейка с комментарием, возвращает комментарий строкой. ЧТо будет, если и комментарий, и формула - то я не знаю"
  (perga-implementation:perga
   (let комментарий nil)
   (let список-с-комментарием
     (find-if 
      (lambda (elt)
        (and (consp (car elt))
             (equal "annotation" (caar elt))))
      ячейка))
   (when список-с-комментарием
     (setq комментарий 
           (очисть-абзац-от-формата 
            список-с-комментарием
            :комментарий t))
     (assert (stringp комментарий))
     комментарий)))




(print
 (первый-комментарий '(("table-cell" . "неважно1") (("value-type" "string") ("value-type" "string"))
  (("annotation" . "urn:oasis:names:tc:opendocument:xmlns:office:1.0")
   (("caption-point-y" "-8.11mm") ("caption-point-x" "-4.07mm") ("y" "30.85mm")
    ("x" "420.05mm") ("height" "9.52mm") ("width" "115.34mm")
    ("text-style-name" "P2") ("style-name" "gr5"))
   (("date" . "http://purl.org/dc/elements/1.1/") NIL "2021-03-29T00:00:00")
   (("p" . "неважно2") (("style-name" "P1"))
    (("span" . "неважно2") (("style-name" "T3")) "Церковнославянское \"как\".")))
  (("p" . "неважно2") NIL "аки"))))


(defun текст-ячейки (ячейка)
  "Не сработает для формулы"
  (perga-implementation:perga
   (let список-с-текстом
     (find-if 
      (lambda (elt)
        (and (consp (car elt))
             (equal "p" (caar elt))))
      ячейка))
   (when список-с-текстом
     (let текст (очисть-абзац-от-формата список-с-текстом))
     ; (let текст (third список-с-текстом))
     (assert (stringp текст))
     текст)))

(assert
 (string= 
  "aka"
  (текст-ячейки 
   '(("table-cell" . неважно) (("value-type" "string") ("value-type" "string"))
     (("p" . "urn:oasis:names:tc:opendocument:xmlns:text:1.0") NIL "aka")))))



(assert '(string= (значение-атрибута-тега '(("table-cell" . "urn:oasis:names:tc:opendocument:xmlns:table:1.0")
  (("value-type" "string") ("value-type" "string")
   ("number-columns-repeated" "2"))
  (("p" . "urn:oasis:names:tc:opendocument:xmlns:text:1.0") NIL "ПРАВЬМЯ"))
                                          "table-cell"
                                          "number-columns-repeated")
         "2"))
                                                                   

(defun содержимое-ячейки (ячейка)
  "Одна ячейка в XML формате может порождать от 0 до Эн ячеек, 
поэтому мы возвращаем список из порождённых ячеек"
  (perga-implementation:perga
   (let к-во-повторов (значение-атрибута-тега ячейка nil "number-columns-repeated"))
   (when к-во-повторов
     (setq к-во-повторов (parse-integer к-во-повторов)))
   (flet повтори-ячейку (дя)
     (let рез nil)
     (dotimes (и (or к-во-повторов 1))
       (push (COPY-Дя дя) рез))
     рез)
   (let имя-тега (имя-тега ячейка))
   (when (string= имя-тега "covered-table-cell")
     (return-from содержимое-ячейки
                  (повтори-ячейку (make-Дя))))
   (assert (string= имя-тега "table-cell"))
   (:@ multiple-value-bind (url текст-url) (формула-только-url ячейка))
   (when url
     (return-from содержимое-ячейки
                  (повтори-ячейку (MAKE-Дя :Url url :Текст текст-url))))
   (let комментарий (первый-комментарий ячейка))
   (let текст (текст-ячейки ячейка))
   (повтори-ячейку (MAKE-Дя :Текст текст :Комментарий комментарий))))

(defparameter *первый-диалект* 3)
(defparameter *за-концом-диалектов* 17)

(defun содержимое-ячеек-строки (строка)
  "Урезаем до последнего диалекта"
  (perga-implementation:perga
   (let ячейки (nthcdr 2 строка))
   (let рез nil)
   (dolist (я ячейки)
     (setf рез (nconc рез (содержимое-ячейки я))))
   (subseq рез 0 *за-концом-диалектов*)))

(defparameter *строка-с-формулой* (nth 19 *строки*))

;; строка с комментарием нужна нам только для разработки/тестирования
(defparameter *строка-с-комментарием* (nth 3 *строки*))

(print (mapcar 'содержимое-ячеек-строки 
               `(,*строка-с-формулой* ,*строка-с-комментарием*)))

(defparameter *пользователи* 
  (with-open-file (in "c:/promo.yar/google-to-semdict/users.lisp")
    (read in)))


(defun список-диалектов ()
  (perga-implementation:perga
   (let Список-Дя (содержимое-ячеек-строки *строка-имён-языков*))
   (let черновик
     (iterk:iter 
      (:for id :from 0)
      (:for дя :in Список-Дя)
      (assert (null (Дя-Комментарий дя)))
      (assert (null (Дя-Url дя)))
      (unless 
          (member (Дя-Текст дя)
                  '("Англ слово (сочетание)" "Темы" "Смысл" "Обсуждение" "Далее идут ссылки" "Алексей Дроздов" NIL) :test 'equal)
        (:collect
         (MAKE-Диалект :Id id :Slug (Дя-Текст дя) :Commentary (Дя-Текст дя))))))
   (dolist (ди черновик)
     (cond
      ((equal (Диалект-Slug ди) "budden (яп \"Яр\")")
       (setf (Диалект-Slug ди) "Яр"))
      ((equal (Диалект-Slug ди) "Другие авторы")
       (setf (Диалект-Slug ди) "Другие-авторы"))
      ((equal (Диалект-Slug ди) "Официальный перевод")
       (setf (Диалект-Slug ди) "Популярные-переводы")
       (setf (Диалект-Commentary ди) "Популярные-переводы"))))
   черновик))

(defparameter *диалекты*
  (список-диалектов))

(defun назначь-владельцев-диалектов ()
  (flet ((f1 (d-id o-id)
           (let ((d (find d-id *диалекты* :test '= :key 'Диалект-Id)))
             (assert d)
             (setf (Диалект-Ownerid d) o-id))))
    (f1 6 3)
    (f1 7 3)
    (f1 8 6)
    (f1 9 4)
    (f1 10 7)))

(назначь-владельцев-диалектов)

(defun форматировать-для-insert (данное)
  (cond
   ((null данное) 'null)
   ((stringp данное)
    (with-output-to-string (п)
      (princ "'" п)
      (iterk:iter
       (:for б :in-vector данное)
       (case б
         (#\' (princ "''" п))
         (t (princ б п))))
      (princ "'" п)))
   ((integerp данное)
    (prin1-to-string данное))
   (t
    (error "Не умею такое напечатать для insert"))))

(defun команда-вставки-пользователя (поль п)
  (format п "~&insert into sduser (id, nickname, registrationemail, salt, hash, registrationtimestamp)
values (~A, ~A, ~A, '', '', current_timestamp);~%"
          (Пользователь-Id поль)
          (форматировать-для-insert (Пользователь-Nickname поль))
          (форматировать-для-insert (Пользователь-Registrationemail поль))
          ))

(defun команды-вставки-пользователей (п)
  (dolist (пол *пользователи*)
    (команда-вставки-пользователя пол п)))


(defun команда-вставки-диалекта (д п)
  (format п "~&insert into tlanguage (id, slug, commentary, ownerid)
values (~A, ~A, ~A, ~A);~%"
          (Диалект-Id д)
          (форматировать-для-insert (Диалект-Slug д))
          (форматировать-для-insert (Диалект-Commentary д))
          (Диалект-Ownerid д)))

(defun команды-вставки-диалектов (п)
  (dolist (д *диалекты*)
    (команда-вставки-диалекта д п)))


(defun команды-вставки-строки (сп номер-строки п)
  (perga-implementation:perga
   (let Senseid (+ номер-строки 1))
   (let сч -1)
   (let oword "")
   (let theme "")
   (let phrase "")
   (let senses nil)
   (dolist (я сп)
     (incf сч)
     (let номер-колонки (+ сч 1))
     (cond
      ((= сч 0) ; слово
       (setf oword (Дя-Текст я))
       (assert (null (Дя-Url я)))
       (assert (null (Дя-Комментарий я))))
      ((= сч 1) ; тема
       (setf theme (Дя-Текст я))
       (assert (null (Дя-Url я)))
       (assert (null (Дя-Комментарий я))))
      ((= сч 2) ; смысл
       (setf phrase (Дя-Текст я))
       (assert (null (Дя-Url я)))
       (assert (null (Дя-Комментарий я))))
      ((and (<= *первый-диалект* сч) (< сч *за-концом-диалектов*))
       (let Word (Дя-Текст я))
       (cond
        ((or
          (and (= номер-строки 17) (= номер-колонки 9)))
         (format t "~&-- пропускаю ячейку ~S~%" я))
        ((null Word)
         (assert (null (Дя-Url я)))
         (assert (null (Дя-Комментарий я))))
        (t
         (let lws (MAKE-Lws :Languageid сч 
                            :Word (Дя-Текст я) 
                            :Senseid Senseid 
                            :Commentary
                            (budden-tools:str++ 
                             (or (Дя-Url я) "")
                             (or (Дя-Комментарий я) ""))))
         (push lws senses))))
      (t ; что за мусор? Не пропустим!
       (error "Чё это"))))
   (format п "~&insert into tsense (id, oword, theme, phrase)
values (~A, ~A, ~A, ~A);~%"
           Senseid
           (форматировать-для-insert oword)
           (форматировать-для-insert theme)
           (форматировать-для-insert phrase)
           )
   (dolist (lws senses)
     (format п "~&insert into tlws (languageid, word, senseid, commentary)
                values (~A, ~A, ~A, ~A);~%"
             (Lws-Languageid lws)
             (форматировать-для-insert (Lws-Word lws))
             Senseid
             (форматировать-для-insert (Lws-Commentary lws))))))


(defun команды-вставки-строк (п)
  (perga-implementation:perga
   (let номер-строки 1)
   (dolist (стр (subseq *строки* *индекс-строки-с-первым-смыслом*))
     (let ячейки (содержимое-ячеек-строки стр))
     (команды-вставки-строки ячейки номер-строки п)
     (incf номер-строки))))

(with-open-file (п "c:/promo.yar/google-to-semdict/final-script.sql"
                   :direction :output
                   :if-exists :supersede)
  (команды-вставки-пользователей п)
  (команды-вставки-диалектов п)
  (команды-вставки-строк п))

