"use client";

import { useEffect, useRef, useState } from "react";
import { isError, UserDTO } from "../../../middleware";
import { WsMessage, WsMessageHandler } from "../lobby/Lobby";
import { toast, ToastContainer } from "react-toastify";
import RoomChat from "./RoomChat";
import { ChatMessageDTO, isChatMessageDTO } from "../Message";
import PlayerTable from "./PlayerTable";
import Question from "./Question";
import Board from "./Board";
import { getAvatar } from "../lobby/LobbyRoom";
import Link from "next/link";
import ControlButtons from "./ControlButtons";
import { useRouter } from "next/navigation";
import { PackPreview } from "../lobby/NewRoomModal";

export type HostDTO = UserDTO & { isConnected: boolean };
export type PlayerDTO = UserDTO & { score: number; isConnected: boolean };
export type BoardQuestion = {
  index: number;
  value: number;
  hasBeenPlayed: boolean;
};

export type RoomDTO = {
  id: string;
  name: string;
  packPreview: PackPreview;
  host: HostDTO | null;
  players: PlayerDTO[];
  currentRound: string | null;
  availableQuestions: { [key: string]: BoardQuestion[] } | null;
  currentPlayer: string | null;
  currentQuestion: HiddenQuestion | Question | null;
  answeringPlayer: string | null;
  finalRoundState: FinalRoundState;
  allowedToAnswer: string[];

  isPaused: boolean;
};

const dummyRoom: RoomDTO = {
  id: "",
  name: "",
  packPreview: { id: "", name: "" },
  players: [],
  host: null,
  currentRound: null,
  availableQuestions: null,
  currentPlayer: null,
  currentQuestion: null,
  answeringPlayer: null,
  allowedToAnswer: [],
  finalRoundState: {
    isActive: false,
    availableQuestions: {},
    question: null,
    bets: null,
  },
  isPaused: false,
};

export const isRoom = (obj: unknown): obj is RoomDTO => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyRoom).every((key) => Object.hasOwn(obj, key));
};

export type HiddenQuestion = {
  index: number;
  value: number;
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

type Question = HiddenQuestion & { answers: string[]; comment: string | null };

export const isQuestion = (obj: unknown, isHost: boolean): obj is Question => {
  return isHost;
};

export type FinalRoundState = {
  isActive: boolean;
  availableQuestions: { [key: string]: boolean } | null;
  question: HiddenFinalRoundQuestion | FinalRoundQuestion | null;
  bets: { playerId: string; amount: number }[] | null;
};

type HiddenFinalRoundQuestion = {
  category: string;
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

type FinalRoundQuestion = HiddenFinalRoundQuestion & {
  answers: string[];
  comment: string | null;
};

export default function Room({
  user,
  initialRoom,
}: {
  user: UserDTO;
  initialRoom: RoomDTO;
}) {
  const router = useRouter();
  const [room, setRoom] = useState(initialRoom);
  const [messages, setMessages] = useState<ChatMessageDTO[]>([
    {
      from: { id: "1", name: "nikita", avatar: null },
      text: "hi everyone",
    },
    {
      from: { id: "4", name: "danya", avatar: null },
      text: "i dont know what im doing here",
    },
    {
      from: { id: "4", name: "danya", avatar: null },
      text: "having fun, hopefully..",
    },
    {
      from: { id: "5", name: "volodymyr", avatar: null },
      text: "lets start the game already",
    },
  ]);

  const wsConn = useRef<WebSocket | null>(null);

  useEffect(() => {
    wsConn.current = new WebSocket(
      `ws://${process.env.NEXT_PUBLIC_BACKEND_HOST}/ws/room/${room.id}`
    );
    const handlers = new Map<string, WsMessageHandler>();

    handlers.set("room", (payload) => {
      if (!isRoom(payload)) return;
      setRoom(payload);
    });
    handlers.set("chat", (payload) => {
      if (!isChatMessageDTO(payload)) return;
      setMessages((messages) => [...messages, payload]);
    });
    handlers.set("error", (payload) => {
      if (!isError(payload)) return;
      toast.error(payload.error, { containerId: "lobby" });
    });

    wsConn.current.addEventListener(
      "message",
      (ev: MessageEvent<WsMessage>) => {
        const handler = handlers.get(ev.data.event);
        if (handler) handler(ev.data.payload);
      }
    );

    wsConn.current.addEventListener("close", () => {
      toast.error("Disconnected from server", { containerId: "room" });
    });

    return () => {
      wsConn.current?.close();
    };
  }, [room.id]);

  const sendMessage = (text: string) => {
    wsConn.current?.send(JSON.stringify({ event: "chat", payload: { text } }));
  };

  const isHost = user.id === room.host?.id;
  console.log(isHost);

  const start = () => {
    if (!isHost) return;
    wsConn.current?.send(JSON.stringify({ event: "start" }));
  }

  const togglePause = () => {
    if (!isHost) return;
    wsConn.current?.send(JSON.stringify({ event: "togglePause" }));
  }

  const leave = () => router.push("/");

  return (
    <>
      <main className="flex flex-col-reverse md:flex-row gap-2 flex-1 min-w-0 min-h-0 p-2">
        <div className="flex-[3_1_0%] flex flex-col gap-2 min-w-0 min-h-0">
          <div className="flex-1 w-full rounded surface p-2">
            {room.availableQuestions &&
              (room.currentQuestion ? (
                <Question question={room.currentQuestion as HiddenQuestion} />
              ) : (
                <Board
                  availableQuestions={room.availableQuestions}
                  chooseQuestion={console.log}
                  isCurrentPlayer={user.id === room.currentPlayer}
                />
              ))}
          </div>
          <div
            className={`w-full flex flex-wrap justify-around gap-3 border rounded p-3`}
          >
            {room.players.map((player, index) => (
              <PlayerTable
                key={index}
                player={player}
                isChoosing={!room.currentQuestion && player.id === room.currentPlayer}
                isAnswering={player.id === room.answeringPlayer}
              />
            ))}
          </div>
          <div className="w-full h-14 rounded primary hover:opacity-85"></div>
        </div>
        <div className="flex-1 flex flex-col gap-2">
          <div className="rounded surface">
            <div className="w-full flex p-2">
              <div className="flex-[1_0_auto]">
                <p className="text-lg font-semibold">{room.name}</p>
                <p className="text-sm font-normal">
                  Pack:{" "}
                  <Link
                    className="pack-link"
                    href={`/pack/${room.packPreview.id}`}
                    target="_blank"
                  >
                    {room.packPreview.name}
                  </Link>
                </p>
              </div>
              <div className={`flex-auto max-w-20 flex flex-col justify-center items-center${ room.host?.isConnected ? "" : " opacity-50" }`}>
                <div className="w-10">{getAvatar(room.host)}</div>
                <p
                  className="w-full text-center text-sm truncate text-white"
                  title={room.host?.name}
                >
                  {room.host?.name}
                </p>
              </div>
            </div>
            <ControlButtons
              isHost={isHost}
              isGameStarted={
                !!room.currentRound || room.finalRoundState.isActive
              }
              isPaused={room.isPaused}
              start={start}
              togglePause={togglePause}
              leave={leave}
            />
          </div>
          <RoomChat user={user} messages={messages} sendMessage={sendMessage} />
        </div>
      </main>
      <ToastContainer
        containerId="room"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
