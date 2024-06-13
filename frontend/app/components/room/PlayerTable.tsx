import React from "react";
import { PlayerDTO } from "./Room";
import { getAvatar } from "../lobby/LobbyRoom";

export default function PlayerTable({
  player,
  isChoosing,
  isAnswering,
}: {
  player: PlayerDTO;
  isChoosing: boolean;
  isAnswering: boolean;
}) {
  const borderColor = isChoosing ? "border-yellow-400" : isAnswering ? "border-orange-600" : "border-white";
  const borderWidth = isChoosing || isAnswering ? "border-2" : "border";
  return (
    <div className={`min-w-8 max-w-24 flex-1${ player.isConnected ? "" : " opacity-50" }`}>
      <div
        className={`w-full aspect-square ${borderWidth} ${borderColor}`}
      >
        {getAvatar(player)}
      </div>
      <div
        className={`w-full ${borderWidth} ${borderColor} rounded-b`}
      >
        <p
          className="w-full text-center text-base truncate hidden md:block px-1"
          title={player.name}
        >
          {player.name}
        </p>
        <p className="w-full text-center text-sm font-extrabold truncate">
          {player.score}
        </p>
      </div>
    </div>
  );
}
