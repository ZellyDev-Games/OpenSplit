package bridge

import (
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/skin"
)

type Skin struct {
	runtimeProvider RuntimeProvider

	skinUpdatedCh <-chan string

	skinModelCh <-chan skin.SkinModel

	skinProvider skin.SkinProvider
}

func NewSkin(
	skinUpdateCh <-chan string,
	skinModelCh <-chan skin.SkinModel,
	runtimeProvider RuntimeProvider,
	skinProvider skin.SkinProvider,
) *Skin {

	return &Skin{

		runtimeProvider: runtimeProvider,

		skinUpdatedCh: skinUpdateCh,

		skinModelCh: skinModelCh,

		skinProvider: skinProvider,
	}

}

func (b *Skin) StartUIPump() {

	//
	// Actual disk skin changes.
	//
	go func() {

		for address := range b.skinUpdatedCh {

			b.runtimeProvider.EventsEmit(
				"skin:reload",
				address,
			)

		}

		logger.Debug(
			logModule,
			"skin reload pump stopped",
		)

	}()

	//
	// Parsed editor model.
	//
	go func() {

		for model := range b.skinModelCh {

			b.runtimeProvider.EventsEmit(
				"skin:model",
				model,
			)

		}

		logger.Debug(
			logModule,
			"skin model pump stopped",
		)

	}()

	logger.Debug(
		logModule,
		"skin UI pumps started",
	)

}
