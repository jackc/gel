# GST - Go Static Templates

GST is a templating library that compiles templates into Go functions.

## Installation

    $ gem install gst

## Usage

    $ gst users_index.gst | gofmt > users_index.go

## Example

package: main
func: EscapeHtml
escape: html
---
<p>Hello, <%= "<Jack>" %>!</p>


## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
