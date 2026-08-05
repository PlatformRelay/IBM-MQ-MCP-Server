package mqadmin_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestValidateRawMQSCAllowsReadOnlyVerbs(t *testing.T) {
	t.Parallel()
	allowed := []string{
		"DISPLAY QLOCAL('DEV.QUEUE.1')",
		"dis qmgr all",
		"Ping QMgr",
		"  display channel('DEV.SVRCONN')",
	}
	for _, cmd := range allowed {
		if err := mqadmin.ValidateRawMQSCCommand(cmd); err != nil {
			t.Fatalf("ValidateRawMQSCCommand(%q) = %v, want nil", cmd, err)
		}
	}
}

func TestValidateRawMQSCDeniesMutatingVerbsBeforeNetwork(t *testing.T) {
	t.Parallel()
	denied := []string{
		"ALTER QLOCAL('DEV.QUEUE.1') MAXDEPTH(1000)",
		"DEFINE QLOCAL('NEW.Q')",
		"DELETE QLOCAL('OLD.Q')",
		"SET AUTHREC PROFILE('DEV.**') OBJTYPE(QUEUE) PRINCIPAL('alice') AUTHADD(ALL)",
		"REFRESH SECURITY TYPE(CONNAUTH)",
	}
	for _, cmd := range denied {
		err := mqadmin.ValidateRawMQSCCommand(cmd)
		if err == nil {
			t.Fatalf("ValidateRawMQSCCommand(%q) = nil, want deny", cmd)
		}
		if !errors.Is(err, mqadmin.ErrMQSCVerbDenied) {
			t.Fatalf("ValidateRawMQSCCommand(%q) = %v, want ErrMQSCVerbDenied", cmd, err)
		}
	}
}

func TestValidateRawMQSCRejectsEmptyAndMultiCommand(t *testing.T) {
	t.Parallel()
	cases := map[string]error{
		"":                                    mqadmin.ErrMQSCCommandEmpty,
		"   \t":                               mqadmin.ErrMQSCCommandEmpty,
		"DISPLAY QLOCAL(A); DELETE QLOCAL(A)": mqadmin.ErrMQSCMultipleCommands,
	}
	for cmd, want := range cases {
		err := mqadmin.ValidateRawMQSCCommand(cmd)
		if !errors.Is(err, want) {
			t.Fatalf("ValidateRawMQSCCommand(%q) = %v, want %v", cmd, err, want)
		}
	}
}

func TestValidateRawMQSCRejectsNewlineStatementInjection(t *testing.T) {
	t.Parallel()
	cases := []string{
		"DISPLAY QLOCAL('X')\nALTER QLOCAL('X') MAXDEPTH(1000)",
		"DISPLAY QLOCAL('X')\r\nALTER QLOCAL('X') MAXDEPTH(1000)",
		"DISPLAY QLOCAL('X')\rALTER QLOCAL('X') MAXDEPTH(1000)",
		"  DISPLAY QLOCAL('X') \n\tALTER QLOCAL('X') MAXDEPTH(1000)",
	}
	for _, cmd := range cases {
		err := mqadmin.ValidateRawMQSCCommand(cmd)
		if !errors.Is(err, mqadmin.ErrMQSCMultipleCommands) {
			t.Fatalf("ValidateRawMQSCCommand(%q) = %v, want ErrMQSCMultipleCommands", cmd, err)
		}
	}
}

func TestRedactMQSCCommandText(t *testing.T) {
	t.Parallel()
	in := "DISPLAY QLOCAL('DEV') WHERE password=secret-value"
	out := mqadmin.RedactMQSCCommandText(in)
	if strings.Contains(out, "secret-value") {
		t.Fatalf("redacted = %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("redacted = %q, want placeholder", out)
	}
}
