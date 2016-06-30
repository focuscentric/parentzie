
function makeWYSIWYG(editor) {
  //If the DOM element we want to edit exists
  if (editor) {
    //We create the buttons container
    var buttons_container = document.createElement('div');

    //We define some properties to it...
    buttons_container.style.textAlign = 'center';
    buttons_container.style.marginTop = '5px';
    buttons_container.className = 'makeWYSIWYG_buttons_container';

    //We create the buttons inside the container
    buttons_container.innerHTML = '' +
    '<button class="makeWYSIWYG_editButton">Edit</button>' +
    '<button class="makeWYSIWYG_viewHTML">HTML</button>' +
    '<button class="makeWYSIWYG_addBlock" data-elm="p">P</button>' +
    '<button class="makeWYSIWYG_addBlock" data-elm="ol">OL</button>' +
    '<div class="makeWYSIWYG_buttons" style="display: none;">' +
    '<button data-tag="bold"><b>Bold</b></button>' +
    '<button data-tag="italic"><em>Italic</em></button>' +
    '<button data-tag="underline"><ins>Underline</ins></button>' +
    '<button data-tag="strikeThrough"><del>Strike</del></button>' +
    '<button data-tag="insertUnorderedList">&bull; Unordered List</button>' +
    '<button data-tag="insertOrderedList">1. Ordered List</button>' +
    '<button data-tag="createLink"><ins style="color: blue;">Link</ins></button>' +
    '<button data-tag="insertImage">Image</button>' +
    '<button data-value="h1" data-tag="heading">Main title</button>' +
    '<button data-value="h2" data-tag="heading">Subtitle</button><br />' +
    '<button data-tag="removeFormat">Remove format</button>' +
    '</div>';

    //We insert the buttons after the editor
    var parent = editor.parentNode;

    if (parent.lastchild == editor) {
      parent.appendChild(buttons_container);
    }
    else {
      parent.insertBefore(buttons_container, editor.nextSibling);
    }

    editor.isEditable = false; //By default, the element is not editable
    editor.setAttribute('contenteditable', false);

    //This function permits to make the element editable or not
    editor.makeEditable = function (bool) {
      //Protect the value
      bool = bool == undefined ? true : (typeof bool === 'boolean' ? bool : true);

      //Change the editable state
      this.isEditable = bool;
      this.setAttribute('contenteditable', bool);

      //Show/Hide the buttons
      if (bool) {
        buttons_container.querySelector('.makeWYSIWYG_buttons').style.display = 'block';
      }
      else {
        buttons_container.querySelector('.makeWYSIWYG_buttons').style.display = 'none';
      }
    };

    //Click on the "Edit" button
    buttons_container.querySelector('.makeWYSIWYG_editButton').addEventListener('click', function (e) {
      if (editor.isEditable) {
        editor.makeEditable(false);
        this.innerHTML = 'Edit';
      }
      else {
        editor.makeEditable(true);
        this.innerHTML = 'Save';
      }
      e.preventDefault();
    }, false);

    //Click on the "View HTML" button
    buttons_container.querySelector('.makeWYSIWYG_viewHTML').addEventListener('click', function (e) {
      alert(editor.innerHTML);
      e.preventDefault();
    }, false);

    //Click for adding new block
    var addElm = buttons_container.querySelectorAll('.makeWYSIWYG_addBlock');

    for (var i = 0; i < addElm.length; i++) {

      addElm[i].addEventListener('click', function (e) {
        var newP = document.createElement('p');
        newP.setAttribute('id', guid());
        newP.classList.add('para');
        console.log(this.getAttribute('data-elm'));
        switch (this.getAttribute('data-elm')) {
          case 'p':
            newP.innerText = 'Nouveau paragraphe';
            break;
          case 'ol':
            newP.innerHTML = '<ol><li>text ici</li><li>text ici</li></ol>';
            break;
        }

        if (parent.lastchild == buttons_container)
          parent.appendChild(newP);
        else
          parent.insertBefore(newP, buttons_container.nextSibling)
      }, false);
    }

    //Get the format buttons
    var buttons = buttons_container.querySelectorAll('button[data-tag]');

    //For each of them...
    for (var i = 0, l = buttons.length; i < l; i++) {
      //We bind the click event
      buttons[i].addEventListener('click', function (e) {
        var tag = this.getAttribute('data-tag');
        switch (tag) {
          case 'createLink':
            var link = prompt('Please specify the link.');
            if (link) {
              document.execCommand('createLink', false, link);
            }
            break;

          case 'insertImage':
            var src = prompt('Please specify the link of the image.');
            if (src) {
              document.execCommand('insertImage', false, src);
            }
            break;

          case 'heading':
            try {
              document.execCommand(tag, false, this.getAttribute('data-value'));
            }
            catch (e) {
              //The browser doesn't support "heading" command, we use an alternative
              document.execCommand('formatBlock', false, '<' + this.getAttribute('data-value') + '>');
            }
            break;

          default:
            document.execCommand(tag, false, this.getAttribute('data-value'));
        }
        e.preventDefault();
      });
    }
  }
  return editor;
};

function guid() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

$.fn.insertAtCaret = function (text) {
  return this.each(function () {
    if (document.selection && this.tagName == 'TEXTAREA') {
      //IE textarea support
      this.focus();
      sel = document.selection.createRange();
      sel.text = text;
      this.focus();
    } else if (this.selectionStart || this.selectionStart == '0') {
      //MOZILLA/NETSCAPE support
      startPos = this.selectionStart;
      endPos = this.selectionEnd;
      scrollTop = this.scrollTop;
      this.value = this.value.substring(0, startPos) + text + this.value.substring(endPos, this.value.length);
      this.focus();
      this.selectionStart = startPos + text.length;
      this.selectionEnd = startPos + text.length;
      this.scrollTop = scrollTop;
    } else {
      // IE input[type=text] and other browsers
      this.value += text;
      this.focus();
      this.value = this.value;    // forces cursor to end
    }
  });
};

$(function () {
  /*
  $('#editor p').each(function (i) {
      $(this).addClass('para').attr('id', guid());
  });

  $('#editor').on({
      click: function (e) {
          e.preventDefault();

          $(this).removeClass('para');
          makeWYSIWYG(document.getElementById($(this).attr('id')));
      }
  }, '.para');
  */

  $('#preview').hide();


  $('#toolbar a').mousedown(function (e) {
    e.preventDefault();

    var editor = $('#editor');

    var elm = $(this).data('elm');
    switch (elm) {
      case 'h2':
        editor.insertAtCaret('<h2>Titre ici</h2>\n\n');
        break;
      case 'h3':
        editor.insertAtCaret('<h3>Sous-titre ici</h3>\n\n');
        break;
      case 'h4':
        editor.insertAtCaret('<h4>Petit titre ici</h4>\n\n');
        break;
      case 'strong':
        editor.insertAtCaret('<strong>text en gras</strong>');
        break;
      case 'em':
        editor.insertAtCaret('<em>text en gras</em>');
        break;
      case 'p':
        editor.insertAtCaret('\n<p>\nText ici\n</p>\n\n');
        break;
      case 'pil':
        editor.insertAtCaret('\n<p>\n<img class="img_left" src="http://amazon.com/img" ' +
          'alt="Text pour image" />\n\nText ici\n</p>\n\n');
        break;
      case 'pid':
        editor.insertAtCaret('\n<p>\n<img class="img_right" src="http://amazon.com/img" ' +
          'alt="Text pour image" />\n\nText ici\n</p>\n\n');
        break;
      case 'ul':
        editor.insertAtCaret('\n<ul>\n  <li>Item ici</li>\n  <li>Item ici</li>\n</ul>\n\n')
        break;
      case 'ol':
        editor.insertAtCaret('\n<ol>\n  <li>Item ici</li>\n  <li>Item ici</li>\n</ol>\n\n')
        break;
      case 'q':
        editor.insertAtCaret('\n<blockquote>Citation ici</blockquote>\n\n');
        break;
      case 't':
        editor.insertAtCaret('\n<table>\n<thead>\n<tr>\n  <th>Entete 1</th>\n  <th>Entete 2</th>\n</tr>\n</thead>\n' +
          '<tbody>\n<tr>\n  <td>Row 1 Cell 1</td>\n  <td>Row 1 Cell 2</td>\n' +
          '<tr>\n  <td>Row 2 Cell 1</td>\n  <td>Row 2 Cell 2</td>\n</tr>\n</tbody>\n</table>');
        break;
      case 'a':
        editor.insertAtCaret('<a href="http://le-lien-ici.com" title="Titre du lien">Text du lien</a>');
        break;
      case 'img':
        editor.insertAtCaret('<img src="http://amazon.com/img" alt="Text de image" />');
        break;
      case 'preview':
        if ($(this).data('preview') == '0') {
          $('#preview').html($('#editor').val());
          $(this).data('preview', '1').text('edit');
          $('#editor').hide();
          $('#preview').show();
        } else {
          $(this).data('preview', '0').text('preview');
          $('#editor').show();
          $('#preview').hide();
        }


    }
  });

  $('#save').click(function (e) {
    e.preventDefault();

    var tag = $('input[name="tag"]:checked').val();
    var id = parseInt($('#pageid').val());

    var posted = {
      id: id,
      slug: $('#slug').val(),
      tag: tag,
      category: $('#category :selected').val(),
      title: $('#title').val(),
      content: $('#editor').val(),
      author: $('#author :selected').val(),
      pubDate: new Date($('#pub-date').val()),
      thumbnail: $('#img-url').val(),
      isFree: $('#free').prop('checked'),
      isFeatured: $('#featured').prop('checked'),
      hasAudio: $('#audio').prop('checked'),
      hasVideo: $('#video').prop('checked')
    }

    console.log(posted);

    $.ajax({
      type: 'POST',
      url: '/save/' + id + document.location.search,
      dataType: 'json',
      contentType: 'application/json',
      data: JSON.stringify(posted),
      processData: false,
      success: function (data) {
        if (data.state)
          document.location.href = '/article/' + posted.slug;
        else
          alert('Nope, failed... ;(');
      },
      error: function () {
        alert('error');
      }
    });
  });

  $('#delete').click(function (e) {
    e.preventDefault();

    if (confirm('Ceci est irreversible?')) {
      var id = $('#pageid').val()
      document.location.href = '/del/' + id + '?key=wtfhc';
    }
  });
});
