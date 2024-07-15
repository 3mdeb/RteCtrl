RTE controller
==============

## General architecture

```
+------+     +-----+       +-------+
|      |     |     |       |       |
| rest |     | web |       | shell |
|      |     |     |       |       |
+------+--+--+-----+       +---+---+
          |                    |
          |                    |
      +---v---+                |
      |       |                |
      |  mux  |                |
      |       |                |
      +---+---+                |
          |                    |
    +-----v----+               |
    |          |       +-------v------+
    |  HTTP    |       |              |
    |  server  +------->  controller  |
    |          |       |              |
    +----------+       +-------+------+
                               |
               +--------------------------------+
               |               |                |
           +---v--+     +------v----+      +----v--+
           |      |     |           |      |       |
           | file +.....+ flashrom  |      | GPIO  |
           |      |     |           |      |       |
           +------+     +-----------+      +-------+
```

## Building

* Clone:

  ```
  mkdir -p ~/go/src/3mdeb && cd ~/go/src/3mdeb
  git clone git@github.com:3mdeb/RteCtrl.git
  cd RteCtrl
  ```

> if $GOPATH is set, change `~/go` to `$GOPATH`

* For host system:

  ```
  make
  ```

* For `arm` system:

  ```
  make build-arm
  ```

* For `amd64` system:

  ```
  make build-amd64
  ```

## Installation

* Install `RteCtrl` binary to `$PATH`

* Install `config/RteCtrl.cfg` to `/etc/RteCtrl/RteCtrl.cfg`

* Install `web` directory to path pointed in `RteCtrl.cfg`:

  ```
  "web_directory" : "web",
  ```

* Adjust path to `flashrom` in `RteCtrl.cfg` if necessary:

  ```
  "flashrom_bin" : "/usr/local/sbin/flashrom",
  ```

## Usage

Required connection:
* connect DUT's power button pin to J11 header pin 9 (GPIO header with OC
buffers)
* connect DUT's reset pin to J11 pin 8 (GPIO header with OC buffers)
* connect DUT's power supply to RTE J13 connector
* connect RTE J12 connector with DUT's DC connector via DC Jack - DC Jack wire
* (only for flashing procedure) connect DUT's SPI header with RTE J7 SPI header

Open in browser: `<RTE_IP>:8000`

![](https://blog.3mdeb.com/img/REST-API.png)

Buttons responsible for power control:

```
Power ON  - sends a pulse on power button pin to power on DUT
Power OFF - ground power button pin on DUT to force it into ACPI S5 state
Reset     - send a pulse on reset pin on DUT
Relay     - toggles the RTE relay on/off deriving the power to DUT
```

Buttons responsible for flashing firmware:

```
Browse... - browse for file to flash DUT's firmware
Upload    - uploads chosen file to RTE (required)
Flash     - starts flashing firmware process
```
