// @author Couchbase <info@couchbase.com>
// @copyright 2014 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"github.com/couchbase/gometa/common"
	"sync"
)

/////////////////////////////////////////////////
// Type Declaration
/////////////////////////////////////////////////

type observer struct {
	packets chan common.Packet
	head    common.Packet

	mutex   sync.Mutex
	isPause bool
	readych chan bool
}

func NewObserver() *observer {

	return &observer{
		packets: make(chan common.Packet, common.MAX_PROPOSALS),
		head:    nil,
		isPause: false,
		readych: make(chan bool, 1)} // buffered - unblock sender
}

func (o *observer) send(msg common.Packet) {

	// Don't need to use this for now
	//o.waitForReady()

	defer common.SafeRun("observer.Send()",
		func() {
			//TODO: handle the case when the channel is full.
			// We don't want send() to block since the caller
			// can be holding mutex.
			o.packets <- msg
		})
}

func (o *observer) pause() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.isPause = true
}

func (o *observer) waitForReady() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.isPause {
		<-o.readych
	}
}

func (o *observer) resume() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.isPause = false
	o.readych <- true
}

func (o *observer) getNext() common.Packet {
	if o.head != nil {
		head := o.head
		o.head = nil
		return head
	}

	if len(o.packets) > 0 {
		packet := <-o.packets
		return packet
	}

	return nil
}

func (o *observer) hasData() bool {

	return len(o.packets) != 0
}

func (o *observer) peekFirst() common.Packet {
	if o.head != nil {
		return o.head
	}

	if len(o.packets) > 0 {
		o.head = <-o.packets
		return o.head
	}

	return nil
}
