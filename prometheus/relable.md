# relabel

## label与relabel应用

   

| relabel_configs action | 修改对象| 说明    |
| :-----| :---- | :---- | 
|replace   | label|根据`regex`来去匹配`source_labels`标签上的值，并将改写到`target_label`中标签 | 
|keep     | target |根据`regex`来去匹配`source_labels`标签上的值，如果匹配成功，则采集此`target`,否则不采集 | 
|drop	    | target |根据`regex`来去匹配`source_labels`标签上的值，如果匹配成功，则不采集此`target`,用于排除，与keep相反|
|labeldrop	||使用regex表达式匹配标签，符合规则的标签将从target实例中移除|
|labelkeep|	|使用regex表达式匹配标签，仅收集符合规则的target，不符合匹配规则的不收集|
|labelmap	 | | 根据regex的定义去匹配Target实例所有标签的名称，并且以匹配到的内容为新的标签名称，其值作为新标签的值|

   


## relabel源码解析

