#!/usr/bin/perl

use strict;
use warnings;

use Date::Format;
use File::Slurp;
use Getopt::Compact;
use Log::Dispatch;
use Log::Dispatch::File;
use Net::Ping;
use Proc::Daemon;

sub options_validate {
	my ($opts) = @_;

	for my $i(qw/host log pid/) {
		if (not exists $opts->{$i} or not $opts->{$i}) {
			print "Error: $i is needed\n\n";

			return 0;
		}
	}

	my $ping = Net::Ping->new('icmp');
	if ($ping->ping($opts->{host}) != 1) {
		print "Error: cannot ping host $opts->{host}\n\n";
		return 0;
	}

	return 1;
}

my $options = Getopt::Compact->new(
	name => 'uptime statistic for one host',
	struct => [
		[ 'host', 'The host we want to check', '=s' ],
		[ 'log', 'The log file', '=s' ],
		[ 'pid', 'The pid file', '=s' ],
	]
);

my $opts = $options->opts();

if (not $options->status() or not options_validate($opts)) {
	print $options->usage();

	exit(1);
}

my $pid_file = $opts->{pid};

if (-f $pid_file) {
	my $pid = File::Slurp::read_file($pid_file);

	if ($pid =~ m/(\d+)/sg) {
		print "Already running as pid $1\n";
	}
	else {
		print "No pid in the pid file?!\n";
	}

	exit(0);
}

my $daemon = Proc::Daemon->new();

my $pid = $daemon->Init();

if ($pid) {
	File::Slurp::write_file($pid_file, "$pid\n");

	print "Running with pid $pid\n";
}
else {
	my $keep_running = 1;

	my $log = new Log::Dispatch(
		callbacks => sub {
			my %h = @_;

			return Date::Format::time2str('%Y-%m-%d %H:%M:%S;', time) . $h{message} . "\n";
		}
	);

	$log->add(Log::Dispatch::File->new(
		name => 'file1',
		min_level => 'warning',
		mode => 'append',
		filename => $opts->{log},
	));

	my $time = 0;
	my $seconds_ok = 0;
	my $seconds_notok = 0;
	my $ping = Net::Ping->new('icmp');

	sub log_statistics {
		$log->warning('Uptime of ' . sprintf('%f', $seconds_ok / ($seconds_ok + $seconds_notok) * 100.0) . '% by a total of ' . ($seconds_ok + $seconds_notok) . ' seconds');
	}

	$SIG{HUP} = $SIG{QUIT} = $SIG{TERM} = sub {
		my $sig_name = shift;

		$log->warning("Caught SIG$sig_name: exiting gracefully");

		$keep_running = 0;
	};
	$SIG{INT}  = sub {
		log_statistics();
	};

	my $last_time = time;

	$log->warning("Start pinging $opts->{host}");

	while ($keep_running) {
		my $current_time = time;

		if ($ping->ping($opts->{host}, 1) == 1) {
			$log->warning('1');

			$seconds_ok += $current_time - $last_time;
		}
		else {
			$log->warning('0');

			$seconds_notok += $current_time - $last_time;
		}

		$last_time = $current_time;

		sleep(1);
	}

	log_statistics();

	$daemon->Kill_Daemon();

	unlink($pid_file);
}

exit(0);
