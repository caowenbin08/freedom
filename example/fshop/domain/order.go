package domain

import (
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/infra/transaction"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *Order {
			return &Order{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *Order) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// Order 订单领域服务.
type Order struct {
	Worker       freedom.Worker          //运行时，一个请求绑定一个运行时
	UserRepo     repository.UserRepo     //用户仓库
	OrderRepo    repository.OrderRepo    //订单仓库
	AdminRepo    repository.AdminRepo    //管理仓库
	DeliveryRepo repository.DeliveryRepo //发货仓库
	Transaction  transaction.Transaction //事务组件
}

// Pay 订单支付 .
func (o *Order) Pay(orderNo string, userId int) (e error) {
	cmd := aggregate.NewOrderPayCmd(o.UserRepo, o.OrderRepo, o.Transaction)
	e = cmd.LoadEntity(orderNo, userId)
	if e != nil {
		return
	}

	return cmd.Pay()
}

// Items 订单列表.
func (o *Order) Items(userId int, page, pageSize int) (result []dto.OrderItemRes, totalPage int, e error) {
	items, totalPage, e := o.OrderRepo.Finds(userId, page, pageSize)
	if e != nil {
		return
	}
	for i := 0; i < len(items); i++ {
		item := dto.OrderItemRes{
			OrderNo:    items[i].OrderNo,
			TotalPrice: items[i].TotalPrice,
			Status:     items[i].Status,
		}
		for j := 0; j < len(items[i].Details); j++ {
			goodsItem := struct {
				GoodsId   int    // 商品id
				Num       int    // 数量
				GoodsName string // 商品名称
			}{
				items[i].Details[j].GoodsId,
				items[i].Details[j].Num,
				items[i].Details[j].GoodsName,
			}
			item.GoodsItems = append(item.GoodsItems, goodsItem)
		}

		result = append(result, item)
	}
	return
}

// Delivery 管理员发货服务
func (o *Order) Delivery(req dto.DeliveryReq) (e error) {
	//创建发货聚合根
	cmd := aggregate.NewDeliveryCmd(o.AdminRepo, o.OrderRepo, o.DeliveryRepo, o.Transaction)
	if e = cmd.LoadEntity(req.OrderNo, req.AdminId); e != nil {
		//加载实体失败
		o.Worker.Logger().Error(e)
		return
	}

	//传入快递单号执行命令
	return cmd.Run(req.TrackingNumber)
}
