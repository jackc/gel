require "minitest/autorun"
require "minitest/reporters"
require 'fileutils'
require 'pry'

Minitest::Reporters.use! [Minitest::Reporters::DefaultReporter.new(:color => true)]

$tmpdir = 'test/tmp/' + Time.now.strftime("%Y-%m-%d-%H-%M-%S")
FileUtils.mkdir_p $tmpdir

class GelTest < Minitest::Test
  def gel(args="")
    `./gel #{args}`
  end

  def test_emits_static_content_unchanged
    gel("test/examples/hello_world/hello_world.html > #{$tmpdir}/hello_world.go")
    FileUtils.cp("test/examples/hello_world/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hello_world.go`
    assert_equal "Hello, World!\n", output
  end

  def test_executes_go_code
    gel("test/examples/hey_hey_hey/hey_hey_hey.html > #{$tmpdir}/hey_hey_hey.go")
    FileUtils.cp("test/examples/hey_hey_hey/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hey_hey_hey.go`
    assert_equal "Hey! Hey! Hey! \n", output
  end

  def test_interpolates_strings
    gel("test/examples/string_interpolation/string_interpolation.html > #{$tmpdir}/string_interpolation.go")
    FileUtils.cp("test/examples/string_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go string_interpolation.go`
    assert_equal "Hello, Jack!\n", output
  end

  def test_interpolates_integers
    gel("test/examples/integer_interpolation/integer_interpolation.html > #{$tmpdir}/integer_interpolation.go")
    FileUtils.cp("test/examples/integer_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go integer_interpolation.go`
    assert_equal "1, 2, 3, 4, 5\n", output
  end

  def test_escapes_for_html_by_default_when_escape_html_set
    gel("test/examples/escape_html/escape_html.html > #{$tmpdir}/escape_html.go")
    FileUtils.cp("test/examples/escape_html/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go escape_html.go`
    assert_equal "<p>Hello, &lt;Jack&gt;!</p>\n", output
  end

  def test_includes_parameters_in_function
    gel("test/examples/parameters/parameters.html > #{$tmpdir}/parameters.go")
    FileUtils.cp("test/examples/parameters/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go parameters.go`
    assert_equal "Hello, Jack!\nHello, Jack!\nHello, Jack!\n", output
  end

  def test_includes_imports
    gel("test/examples/imports/imports.html > #{$tmpdir}/imports.go")
    FileUtils.cp("test/examples/imports/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go imports.go`
    assert_equal "1", output
  end

  def test_merges_multiple_templates_into_single_go_file
    gel("test/examples/multiple/escape_html.html test/examples/multiple/hello_world.html > #{$tmpdir}/multiple.go")
    FileUtils.cp("test/examples/multiple/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go multiple.go`
    assert_equal "Hello, World!\n<p>Hello, &lt;Jack&gt;!</p>\n", output
  end
end
