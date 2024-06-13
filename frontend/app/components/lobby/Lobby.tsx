"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { ChatMessageDTO, isChatMessageDTO } from "../Message";
import { isError, UserDTO } from "../../../middleware";
import RoomsList from "./RoomsList";
import NewRoomModal from "./NewRoomModal";
import { toast, ToastContainer } from "react-toastify";
import PasswordModal from "./PasswordModal";
import { isLobbyRoom, LobbyRoomDTO } from "./LobbyRoom";
import LobbyChat from "./LobbyChat";
import AddButton from "../AddButton";

export type WsMessage = { event: string; payload: unknown };
export type WsMessageHandler = (payload: unknown) => void;

type LobbyRoomDeletedDTO = {
  id: string;
};

const dummyLobbyRoomDeleted: LobbyRoomDeletedDTO = {
  id: "1",
};

export const isLobbyRoomDeleted = (
  obj: unknown
): obj is LobbyRoomDeletedDTO => {
  if (typeof obj !== "object" || obj === null) return false;
  return Object.keys(dummyLobbyRoomDeleted).every((key) =>
    Object.hasOwn(obj, key)
  );
};

export default function Lobby({
  user,
  initialRooms,
}: {
  user: UserDTO;
  initialRooms: LobbyRoomDTO[];
}) {
  const [rooms, setRooms] = useState(initialRooms);
  const [messages, setMessages] = useState<ChatMessageDTO[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [isNewRoomModalOpen, setIsNewRoomModalOpen] = useState(false);

  const [passwordModal, setPasswordModal] = useState<{
    roomId: string | undefined;
    isOpen: boolean;
  }>({
    roomId: undefined,
    isOpen: false,
  });

  const wsConn = useRef<WebSocket | null>(null);

  useEffect(() => {
    wsConn.current = new WebSocket(
      `ws://${process.env.NEXT_PUBLIC_BACKEND_HOST}/ws/lobby`
    );
    const handlers = new Map<string, WsMessageHandler>();

    handlers.set("lobby-room", (payload) => {
      if (!isLobbyRoom(payload)) return;
      setRooms((rooms) =>
        rooms.some((room) => room.id === payload.id)
          ? rooms.map((room) => (room.id === payload.id ? payload : room))
          : [...rooms, payload]
      );
    });
    handlers.set("lobby-room-deleted", (payload) => {
      if (!isLobbyRoomDeleted(payload)) return;
      setRooms((rooms) => rooms.filter((room) => room.id !== payload.id));
    });
    handlers.set("chat", (payload) => {
      if (!isChatMessageDTO(payload)) return;
      setMessages((messages) => [...messages, payload]);
    });
    handlers.set("error", (payload) => {
      if (!isError(payload)) return;
      toast.error(payload.error, { containerId: "lobby" });
    });

    wsConn.current.addEventListener("message", (ev: MessageEvent<string>) => {
      const message: WsMessage = JSON.parse(ev.data);
      const handler = handlers.get(message.event);
      if (handler) handler(message.payload);
    });

    wsConn.current.addEventListener("close", () => {
      toast.error("Disconnected from server", { containerId: "lobby" });
    });

    return () => {
      wsConn.current?.close();
    };
  }, []);

  const filteredRooms = useMemo(
    () =>
      rooms.filter((room) =>
        room.name.toLowerCase().includes(searchInput.trim().toLowerCase())
      ),
    [rooms, searchInput]
  );

  const sendMessage = (text: string) => {
    wsConn.current?.send(JSON.stringify({ event: "chat", payload: { text } }));
  };

  return (
    <>
      <main
        className={`flex flex-col sm:flex-row flex-1 gap-3 min-w-0 min-h-0 
        p-3`}
      >
        <div
          className={`flex flex-col min-w-0 min-h-0 relative 
          max-h-[50%] sm:flex-[1_0_0%] sm:max-h-none`}
        >
          <div
            className={`flex items-center min-h-12 border rounded p-2 
            surface`}
          >
            <input
              className="search-room w-full rounded-lg p-1 text-black"
              placeholder="Search existing rooms"
              value={searchInput}
              onChange={(ev) => setSearchInput(ev.target.value)}
            />
          </div>
          <RoomsList
            rooms={filteredRooms}
            openPasswordModal={(roomId: string) =>
              setPasswordModal({ roomId, isOpen: true })
            }
          />
          <AddButton onClick={() => setIsNewRoomModalOpen(true)} />
        </div>
        <LobbyChat
          user={user}
          messages={messages}
          sendMessage={sendMessage}
        />
      </main>
      <NewRoomModal
        isOpen={isNewRoomModalOpen}
        close={() => setIsNewRoomModalOpen(false)}
      />
      <PasswordModal
        isOpen={passwordModal.isOpen}
        close={() => setPasswordModal({ ...passwordModal, isOpen: false })}
        roomId={passwordModal.roomId}
      />
      <ToastContainer
        containerId="lobby"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
