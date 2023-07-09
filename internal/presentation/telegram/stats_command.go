package telegram

//func (r *Presentation) statsCommandHandler(
//	ctx context.Context,
//	entities *tg.Entities,
//	update message.AnswerableMessageUpdate,
//	m *tg.Message,
//) error {
//	peer := utils.GetPeer(entities)
//
//	err := query.Messages(r.telegramApi).
//		GetHistory(peer).
//		BatchSize(100).
//		ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
//			elemMessage, ok := elem.Msg.(*tg.Message)
//			if !ok {
//				return nil
//			}
//			println(elemMessage.GetMessage())
//			println(time.Unix(int64(elemMessage.Date), 0).String())
//			//for key, v := range elem.Entities.Users() {
//			//	println(key)
//			//	utils.LogInterface(v.Username)
//			//}
//			return nil
//		})
//	if err != nil {
//		return err
//	}
//	_, err = r.telegramSender.Reply(*entities, update).
//		StyledText(ctx, html.String(nil, "aaa"))
//	return err
//}
