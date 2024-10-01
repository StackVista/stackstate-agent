from typing import Tuple, Union

SingleComponentKey = str
RepeatedComponentKey = Tuple[SingleComponentKey, int]
ComponentKey = Union[SingleComponentKey, RepeatedComponentKey]

SingleRelationKey = Tuple[ComponentKey, ComponentKey]
RepeatedRelationKey = Tuple[SingleRelationKey, int]
RelationKey = Union[SingleRelationKey, RepeatedRelationKey]
