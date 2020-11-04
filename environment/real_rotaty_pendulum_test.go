package environment

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewEncodedReceiveData(t *testing.T) {
	tests := []struct {
		input string
		data  *EncodedReceiveData
		ok    bool
	}{
		{
			"013J0003F33DA\n",
			&EncodedReceiveData{
				[4]byte{'0', '1', '3', 'J'},
				[3]byte{'0', '0', '0'},
				[2]byte{'3', 'F'},
				[3]byte{'3', '3', 'D'},
				'A',
				'\n',
			},
			true,
		},
		{
			"0D^;9@h3G33CA\n",
			&EncodedReceiveData{
				[4]byte{'0', 'D', '^', ';'},
				[3]byte{'9', '@', 'h'},
				[2]byte{'3', 'G'},
				[3]byte{'3', '3', 'C'},
				'A',
				'\n',
			},
			true,
		},
		{
			"0No@9@g:N33Cn\n",
			&EncodedReceiveData{
				[4]byte{'0', 'N', 'o', '@'},
				[3]byte{'9', '@', 'g'},
				[2]byte{':', 'N'},
				[3]byte{'3', '3', 'C'},
				'n',
				'\n',
			},
			true,
		},
		{
			"0No@9@g:N33Cn\a",
			nil,
			false,
		},
	}

	for i, test := range tests {
		buf := []byte(test.input)
		data, err := NewEncodedReceiveData(buf)
		if test.ok {
			if err != nil {
				t.Errorf("[%d] ok buf failed: %v", i, err)
				continue
			}
			if !reflect.DeepEqual(test.data, data) {
				t.Errorf("[%d] expected %v but %v", i, test.data, data)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("[%d] expected fail but err is nil", i)
				continue
			}
		}
	}
}

func TestEncodedReceiveData_ToReceiveData(t *testing.T) {
	tests := []struct {
		input    string
		expected *ReceiveData
	}{
		{
			"013J0003F33DA\n",
			&ReceiveData{
				4314,
				0,
				214,
				12500,
			},
		},
		{
			"0D^;9@h3G33CA\n",
			&ReceiveData{
				84875,
				37944,
				215,
				12499,
			},
		},
	}

	for i, test := range tests {
		buf := []byte(test.input)
		encData, err := NewEncodedReceiveData(buf)
		if err != nil {
			t.Errorf("[%d] failed NewEncodedReceiveData: %v: %v", i, test.input, err)
			continue
		}

		receiveData, err := encData.ToReceiveData()
		if test.expected == nil {
			if err != nil {
				t.Errorf("[%d] expected err but err is nil: %q", i, test.input)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("[%d] got error: %v", i, err)
				continue
			}
			if !reflect.DeepEqual(test.expected, receiveData) {
				t.Errorf("[%d] expected %v, got %v", i, test.expected, receiveData)
			}
		}
	}
}

func TestReceiveData_rawEncoderToRad(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{
			0,
			"0.00",
		},
		{
			50000,
			"1.57",
		},
		{
			175000,
			"-0.79",
		},
	}

	for i, test := range tests {
		rad := new(ReceiveData).rawEncoderToRad(test.input)
		str := fmt.Sprintf("%.2f", rad)
		if test.expected != str {
			t.Errorf("[%d] expected %q, but %q", i, test.expected, str)
		}
	}
}

func TestReceiveData_rawPotentiometerToRad(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{
			0,
			"0.00",
		},
		{
			256,
			"1.57",
		},
		{
			896,
			"-0.79",
		},
	}

	for i, test := range tests {
		rad := new(ReceiveData).rawPotentiomaterToRad(test.input)
		str := fmt.Sprintf("%.2f", rad)
		if test.expected != str {
			t.Errorf("[%d] expected %q, but %q", i, test.expected, str)
		}
	}
}

func TestReceiveData_rawPWMDutyToVoltage(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{
			12499,
			"0.00",
		},
		{
			0,
			"5.00",
		},
		{
			0 + 0x10000,
			"-5.00",
		},
		{
			6250 + 0x10000,
			"-2.50",
		},
	}

	for i, test := range tests {
		rad := new(ReceiveData).rawPWMDutyToVoltage(test.input)
		str := fmt.Sprintf("%.2f", rad)
		if test.expected != str {
			t.Errorf("[%d] expected %q, but %q", i, test.expected, str)
		}
	}
}
