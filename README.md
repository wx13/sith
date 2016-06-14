Sith
====

Sith is a console-based text editor written in go. Aside from the normal
text-editor stuff, it has a few notible features:

- multiple cursors
  - edit multiple lines at once
  - indent or comment blocks of code
  - align/unalign text on multple lines.
- automatic indentation detection
- copy/paste history

![screenshot](http://www.wx13.com/sithscreenshot.png)

It is still in "beta", but is pretty stable.  I use it for all my text editing needs.


Binary Install
--------------

Download the latest binary for your platform <https://github.com/wx13/sith/releases/latest>
and just execute it

    ./sith <list of files>


Build from source
-----------------

    cd sith/
    go get .
    go build


Rationale
---------

I was dissatisfied with existing console text editors, so I
wrote one in ruby (<https://github.com/wx13/jedi>).  While that served my
needs for a couple of years, I learned a lot about text editors.  Sith
is my attempt to get things right.

Should you use sith as your text editor?  Heck no!  You should write your own!
But feel free to take a look at the sith code for ideas.


License
-------

MIT license.  See LICENSE.

