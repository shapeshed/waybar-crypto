# WayBar Crypto

![WayBar Crypto][3]

A custom [WayBar][1] module to display market information on a cryptocurrency
using the [Binance API][2].

## Installation

- Download a release for your platform from the [releases][4] page
- Move the binary into your path e.g `/usr/bin/local`
- Create a toml file at `~/.config/waybar-crypto/config.toml`
- Populate the config file with your variables. Toggle features depending on
  what you want to see.

```toml
# Binance API Credentials
api_key = "YOUR_BIANCE_API_KEY"
secret_key = "YOUR_BINANCE_SECRET_KEY"

# Market Ticker
ticker = "BTCUSDT"

# Feature Toggles
show_funding_rate = true
show_open_interest = true
show_volume_change = true
show_long_short_ratio = true

# Custom Colors
color_positive = "#a6e3a1"  # Green
color_negative = "#f38ba8"  # Red
```

Check the installation by running `waybar-crypto` from the command line and you
should see some json output

```sh
waybar-crypto
```

```json
{"text":" $83400.01 \u003cspan color='#a6e3a1'\u003e0.37%\u003c/span\u003e | Funding: \u003cspan color='#a6e3a1'\u003e0.0060%\u003c/span\u003e | OI: \u003cspan color='#a6e3a1'\u003e0.31%\u003c/span\u003e | Vol: \u003cspan color='#a6e3a1'\u003e11.48%\u003c/span\u003e | LSR: \u003cspan color='#a6e3a1'\u003e1.76\u003c/span\u003e","tooltip":"Price Change: 0.37% | Funding Rate: 0.0060% | Open Interest Change: 0.31% | Volume Change: 11.48% | Long-Short Ratio: 1.76"}
```

### WayBar configuration

Modify `~/.config/waybar/config` and add the custom module.

```json
"custom/waybar-crypto": {
  "exec": "/usr/local/bin/waybar-crypto",
  "interval": 60,
  "return-type": "json"
}
```

Choose where to display the module. Here it is displayed in the middle.

```json
"modules-center": [
  "custom/btc",
],
```

[1]: https://github.com/Alexays/Waybar
[2]: https://www.binance.com/binance-api
[3]: assets/waybar-crypto.png "Screenshot showing waybar-crypto"
[4]: https://github.com/shapeshed/waybar-crypto/releases
