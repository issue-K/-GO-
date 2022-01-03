package gee

import "strings"

/*存储路由的结构.普通的map[string]handlefunc只能查找静态路由
这里使用trie实现一个动态路由,支持:name 和 *两种匹配规则
*/
type node struct{
	pattern string //待匹配路由(从根节点到本节点的路由路径,形如/p/:name)
	part string //本节点中存储路由的一部分,如:name
	children []*node //子节点
	isWild bool //part中是否含有:和*,有则为true
}

//查找子节点中第一个匹配成功的节点
func (n *node) matchChild(part string) *node{
	for _,child := range n.children{
		if child.part == part || child.isWild{  //路由完全吻合或路由中存在通配符:,*
			return child
		}
	}
	return nil
}
//查找所有匹配成功的节点,返回节点数组
func (n *node) matchChildren(part string) []*node{
	nodes := make([]*node,0)
	for _,child:= range n.children{
		if child.part == part || child.isWild{
			nodes = append( nodes,child )
		}
	}
	return nodes
}
/* insert
pattern表示需要插入的路由路径,形如/p/:name
parts就是把pattern根据/拆开来形成的数组
dep表示当前在trie树的深度
*/
func (n *node) insert( pattern string,parts []string,dep int){
	if len(parts) == dep{ ///匹配成功
		n.pattern = pattern
		return
	}
	part := parts[dep] //在下一层需要匹配的字符串
	child := n.matchChild(part)
	if child == nil{ //没有该子节点,就新建一个
		child = &node{ part:part,isWild:part[0]==':' || part[0]=='*' }
		n.children = append( n.children,child )
	}
	child.insert( pattern,parts,dep+1 )
}
func (n *node) search( parts []string,dep int) * node{
	if len(parts)==dep || strings.HasPrefix(n.part,"*"){
		if n.pattern==""{
			return nil
		}
		return n
	}
	part := parts[dep]
	children := n.matchChildren(part)

	for _,child := range children{
		result := child.search( parts,dep+1 )
		if result!=nil{
			return result
		}
	}
	return nil
}