#!/usr/bin/perl

use strict;
use warnings;

my $ttyName = `ls -l /dev/phone`;

$ttyName =~ s/.*(ttyUSB\d+).*/$1/sg;

my $port = `find /sys/bus/usb/devices/usb*/ | grep $ttyName`;

$port =~ s/.+\/(.+?):.+?\/.+/$1/sg;

print `echo -n '$port' > /sys/bus/usb/drivers/usb/unbind`;

sleep(2);

print `echo -n '$port' > /sys/bus/usb/drivers/usb/bind`;
