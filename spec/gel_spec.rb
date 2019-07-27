require 'rspec'
require 'fileutils'

$tmpdir = 'spec/tmp/' + Time.now.strftime("%Y-%m-%d-%H-%M-%S")
FileUtils.mkdir_p $tmpdir

RSpec.configure do |config|
  def gel(args="")
    `./gel #{args}`
  end
end

describe "gel" do
  it "emits static content unchanged" do
    gel("spec/examples/hello_world/hello_world.html > #{$tmpdir}/hello_world.go")
    FileUtils.cp("spec/examples/hello_world/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hello_world.go`
    output.should eq "Hello, World!\n"
  end

  it "executes go code" do
    gel("spec/examples/hey_hey_hey/hey_hey_hey.html > #{$tmpdir}/hey_hey_hey.go")
    FileUtils.cp("spec/examples/hey_hey_hey/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go hey_hey_hey.go`
    output.should eq "Hey! Hey! Hey! \n"
  end

  it "interpolates strings" do
    gel("spec/examples/string_interpolation/string_interpolation.html > #{$tmpdir}/string_interpolation.go")
    FileUtils.cp("spec/examples/string_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go string_interpolation.go`
    output.should eq "Hello, Jack!\n"
  end

  it "interpolates integers" do
    gel("spec/examples/integer_interpolation/integer_interpolation.html > #{$tmpdir}/integer_interpolation.go")
    FileUtils.cp("spec/examples/integer_interpolation/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go integer_interpolation.go`
    output.should eq "1, 2, 3, 4, 5\n"
  end

  it "escapes for HTML by default when escape html set" do
    gel("spec/examples/escape_html/escape_html.html > #{$tmpdir}/escape_html.go")
    FileUtils.cp("spec/examples/escape_html/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go escape_html.go`
    output.should eq "<p>Hello, &lt;Jack&gt;!</p>\n"
  end

  it "includes parameters in function" do
    gel("spec/examples/parameters/parameters.html > #{$tmpdir}/parameters.go")
    FileUtils.cp("spec/examples/parameters/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go parameters.go`
    output.should eq "Hello, Jack!\nHello, Jack!\nHello, Jack!\n"
  end

  it "includes imports" do
    gel("spec/examples/imports/imports.html > #{$tmpdir}/imports.go")
    FileUtils.cp("spec/examples/imports/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go imports.go`
    output.should eq "1"
  end

  it "merges multiple templates into single go file" do
    gel("spec/examples/multiple/escape_html.html spec/examples/multiple/hello_world.html > #{$tmpdir}/multiple.go")
    FileUtils.cp("spec/examples/multiple/main.go", $tmpdir)
    output = `cd #{$tmpdir}; go run main.go multiple.go`
    output.should eq "Hello, World!\n<p>Hello, &lt;Jack&gt;!</p>\n"
  end
end
