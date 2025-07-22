package resolver

import (
	"thermal/model"
)

// DTSの全要素をmapにまとめる
func CollectElementsByHref(schema *model.XBRLSchema, result map[string]*model.XMLElement) {
	for i, element := range schema.Elements {
		key := schema.Path + "#" + element.Id
		result[key] = &schema.Elements[i]
	}
	for i := range schema.Imports {
		if schema.Imports[i].Schema != nil {
			CollectElementsByHref(schema.Imports[i].Schema, result)
		}
	}
}

// DTSの全ロールタイプをmapにまとめる
func CollectRoleTypesByHref(schema *model.XBRLSchema, result map[string]*model.RoleType) {
	for _, rt := range schema.RoleTypes {
		key := schema.Path + "#" + rt.Id
		result[key] = &rt
	}
	for i := range schema.Imports {
		if schema.Imports[i].Schema != nil {
			CollectRoleTypesByHref(schema.Imports[i].Schema, result)
		}
	}
}
