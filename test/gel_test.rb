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
    gel("< test/examples/hello_world/hello_world.html | goimports > #{$tmpdir}/hello_world.go")
    output = `cd #{$tmpdir}; go run hello_world.go`
    assert_equal "Hello, World!\n", output
  end

  def test_executes_go_code
    gel("< test/examples/hey_hey_hey/hey_hey_hey.html | goimports > #{$tmpdir}/hey_hey_hey.go")
    output = `cd #{$tmpdir}; go run hey_hey_hey.go`
    assert_equal "Hey! Hey! Hey! \n", output
  end

  def test_interpolates_strings
    gel("< test/examples/string_interpolation/string_interpolation.html | goimports > #{$tmpdir}/string_interpolation.go")
    output = `cd #{$tmpdir}; go run string_interpolation.go`
    assert_equal "Hello, Jack!\n", output
  end

  def test_interpolates_integers
    gel("< test/examples/integer_interpolation/integer_interpolation.html | goimports > #{$tmpdir}/integer_interpolation.go")
    output = `cd #{$tmpdir}; go run integer_interpolation.go`
    assert_equal "1, 2, 3, 4, 5\n", output
  end

  def test_escapes_html_by_default
    gel("< test/examples/escape_html/escape_html.html | goimports > #{$tmpdir}/escape_html.go")
    output = `cd #{$tmpdir}; go run escape_html.go`
    assert_equal "<p>Hello, &lt;Jack&gt;!</p>\n", output
  end
end
