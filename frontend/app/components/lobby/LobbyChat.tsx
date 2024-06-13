"use client";

import {
  KeyboardEventHandler,
  useEffect,
  useRef,
  useState,
} from "react";
import Message, { ChatMessageDTO } from "../Message";
import { UserDTO } from "../../../middleware";

export default function LobbyChat({
  user,
  messages,
  sendMessage,
}: {
  user: UserDTO;
  messages: ChatMessageDTO[];
  sendMessage: (text: string) => void;
}) {
  const [input, setInput] = useState("");
  const scrollableRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    scrollableRef.current?.scroll({
      top: scrollableRef.current?.scrollHeight,
      behavior: "smooth",
    });
  }, [messages]);

  const handleSend = () => {
    if (!input) return;
    sendMessage(input);
    setInput("");
  };

  const onInputKeyDown: KeyboardEventHandler = ev => {
    if (ev.key !== "Enter") return;
    handleSend();
  };

  return (
    <div className="flex flex-col min-w-0 min-h-0 rounded surface flex-[1_0_0%] sm:flex-[2_0_0%]">
      <div
        className="flex flex-col flex-1 overflow-x-auto p-2"
        ref={scrollableRef}
      >
        {messages.map((message, index) => {
          const isOwn = user.id === message.from.id;
          const isPrevUserSame =
            message.from.id === messages[index - 1]?.from.id;
          const isNextUserSame =
            message.from.id === messages[index + 1]?.from.id;
          return (
            <Message
              key={index}
              message={message}
              isOwn={isOwn}
              isPrevUserSame={isPrevUserSame}
              isNextUserSame={isNextUserSame}
            />
          );
        })}
      </div>
      <div className="flex flex-row gap-2 min-h-12 border rounded p-2">
        <input
          className="flex-1 rounded-lg p-1 text-black"
          placeholder="Say something to others"
          value={input}
          onChange={ev => setInput(ev.target.value)}
          onKeyDown={onInputKeyDown}
        />
        <button
          className="w-20 primary rounded-lg font-medium"
          onClick={handleSend}
        >
          Send
        </button>
      </div>
    </div>
  );
}
