(in-package :cl-user)

(proclaim '(optimize debug))
(declaim (optimize debug))

;; дя = данные-ячейки
(eval-when (:compile-toplevel :execute)
  (defstruct Дя
    Текст
    Url
    Комментарий))

(defparameter *xml* 
  (with-open-file (s "c:/promo.yar/google-to-semdict/Англо-Русский\ словарь\ терминов\ и\ слов\ для\ включения\ в\ программы\ -\ 2021-03-29.fods")
(xmls:parse s)))

;;; инспектируем *xml* и достаём оттуда 

(defparameter *body* (nth 9 *xml*))

(defparameter *spreadsheet* (nth 2 *body*))

(defparameter *лист-словарь* (nth 4 *spreadsheet*))

;;; Колонки и строки вынимаем руками

(defparameter *колонки*
  (subseq *лист-словарь* 3 16))

(defparameter *строки*
  (subseq *лист-словарь* 16 139))

(defparameter *строка-имён-языков* 
  (nth 0 *строки*))

(defun формула-только-url (ячейка)
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
   (format nil "<a href=\"~A\">~A</a>" url text)))      


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


(defun содержимое-ячейки (ячейка)
  (perga-implementation:perga
   (let url (формула-только-url ячейка))
   (when url
     (return-from содержимое-ячейки
                  (MAKE-Дя :Url url)))
   (let комментарий (первый-комментарий ячейка))
   (let текст (текст-ячейки ячейка))
   (MAKE-Дя :Текст текст :Комментарий комментарий)))

(defun содержимое-ячеек-строки (строка)
  (perga-implementation:perga
   (let ячейки (nthcdr 2 строка))
   (mapcar #'содержимое-ячейки ячейки)))

(defparameter *строка-с-формулой* (nth 19 *строки*))
(defparameter *строка-с-комментарием* (nth 3 *строки*))

(print (mapcar 'содержимое-ячеек-строки 
               `(,*строка-с-формулой* ,*строка-с-комментарием*)))


