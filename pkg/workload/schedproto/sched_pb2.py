# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: sched.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0bsched.proto\x12\x05sched\"\xb7\x01\n\x0cSchedRequest\x12\x10\n\x08schedAlg\x18\x01 \x01(\t\x12\x16\n\x0einvocationName\x18\x02 \x01(\t\x12\x11\n\tbatchsize\x18\x03 \x01(\r\x12\x19\n\x11runtimeInMilliSec\x18\x04 \x01(\r\x12\x12\n\niterations\x18\x05 \x01(\r\x12\x10\n\x08\x64\x65\x61\x64line\x18\x06 \x01(\x05\x12\x13\n\x0bprevReplica\x18\x07 \x01(\r\x12\x14\n\x0c\x61vailableGPU\x18\x08 \x01(\r\"L\n\nSchedReply\x12\x16\n\x0einvocationName\x18\x01 \x03(\t\x12\x0f\n\x07replica\x18\x02 \x03(\r\x12\x15\n\rschedOverhead\x18\x03 \x01(\r2|\n\x08\x45xecutor\x12\x33\n\x07\x45xecute\x12\x13.sched.SchedRequest\x1a\x11.sched.SchedReply\"\x00\x12;\n\rExecuteStream\x12\x13.sched.SchedRequest\x1a\x11.sched.SchedReply\"\x00(\x01\x42\x17Z\x15schedproto/schedprotob\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'sched_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\025schedproto/schedproto'
  _globals['_SCHEDREQUEST']._serialized_start=23
  _globals['_SCHEDREQUEST']._serialized_end=206
  _globals['_SCHEDREPLY']._serialized_start=208
  _globals['_SCHEDREPLY']._serialized_end=284
  _globals['_EXECUTOR']._serialized_start=286
  _globals['_EXECUTOR']._serialized_end=410
# @@protoc_insertion_point(module_scope)
