from dataclasses import dataclass
from service.anime_service import AnimeService
from service.plot_service import PlotService


@dataclass
class Container:
    plot_service: PlotService
    anime_service: AnimeService


def init_combat_container() -> Container:
    plot_service = PlotService()
    anime_service = AnimeService("weights.pth")

    return Container(anime_service=anime_service, plot_service=plot_service)
