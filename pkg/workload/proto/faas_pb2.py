# -*- coding: utf-8 -*-

#  MIT License
#
#  Copyright (c) 2023 EASL and the vHive community
#
#  Permission is hereby granted, free of charge, to any person obtaining a copy
#  of this software and associated documentation files (the "Software"), to deal
#  in the Software without restriction, including without limitation the rights
#  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#  copies of the Software, and to permit persons to whom the Software is
#  furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included in all
#  copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
#  SOFTWARE.

# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: server/faas.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='server/faas.proto',
  package='faas',
  syntax='proto3',
  serialized_options=b'Z!github.com/eth-easl/loader/server',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x11server/faas.proto\x12\x04\x66\x61\x61s\"T\n\x0b\x46\x61\x61sRequest\x12\x0f\n\x07message\x18\x01 \x01(\t\x12\x19\n\x11runtimeInMilliSec\x18\x02 \x01(\r\x12\x19\n\x11memoryInMebiBytes\x18\x03 \x01(\r\"Q\n\tFaasReply\x12\x0f\n\x07message\x18\x01 \x01(\t\x12\x1a\n\x12\x64urationInMicroSec\x18\x02 \x01(\r\x12\x17\n\x0fmemoryUsageInKb\x18\x03 \x01(\r2;\n\x08\x45xecutor\x12/\n\x07\x45xecute\x12\x11.faas.FaasRequest\x1a\x0f.faas.FaasReply\"\x00\x42#Z!github.com/eth-easl/loader/serverb\x06proto3'
)




_FAASREQUEST = _descriptor.Descriptor(
  name='FaasRequest',
  full_name='faas.FaasRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='message', full_name='faas.FaasRequest.message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='runtimeInMilliSec', full_name='faas.FaasRequest.runtimeInMilliSec', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='memoryInMebiBytes', full_name='faas.FaasRequest.memoryInMebiBytes', index=2,
      number=3, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=27,
  serialized_end=111,
)


_FAASREPLY = _descriptor.Descriptor(
  name='FaasReply',
  full_name='faas.FaasReply',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='message', full_name='faas.FaasReply.message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='durationInMicroSec', full_name='faas.FaasReply.durationInMicroSec', index=1,
      number=2, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='memoryUsageInKb', full_name='faas.FaasReply.memoryUsageInKb', index=2,
      number=3, type=13, cpp_type=3, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=113,
  serialized_end=194,
)

DESCRIPTOR.message_types_by_name['FaasRequest'] = _FAASREQUEST
DESCRIPTOR.message_types_by_name['FaasReply'] = _FAASREPLY
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

FaasRequest = _reflection.GeneratedProtocolMessageType('FaasRequest', (_message.Message,), {
  'DESCRIPTOR' : _FAASREQUEST,
  '__module__' : 'server.faas_pb2'
  # @@protoc_insertion_point(class_scope:faas.FaasRequest)
  })
_sym_db.RegisterMessage(FaasRequest)

FaasReply = _reflection.GeneratedProtocolMessageType('FaasReply', (_message.Message,), {
  'DESCRIPTOR' : _FAASREPLY,
  '__module__' : 'server.faas_pb2'
  # @@protoc_insertion_point(class_scope:faas.FaasReply)
  })
_sym_db.RegisterMessage(FaasReply)


DESCRIPTOR._options = None

_EXECUTOR = _descriptor.ServiceDescriptor(
  name='Executor',
  full_name='faas.Executor',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=196,
  serialized_end=255,
  methods=[
  _descriptor.MethodDescriptor(
    name='Execute',
    full_name='faas.Executor.Execute',
    index=0,
    containing_service=None,
    input_type=_FAASREQUEST,
    output_type=_FAASREPLY,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_EXECUTOR)

DESCRIPTOR.services_by_name['Executor'] = _EXECUTOR

# @@protoc_insertion_point(module_scope)
