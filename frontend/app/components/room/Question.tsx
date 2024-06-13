import React from "react";
import { HiddenQuestion } from "./Room";

export default function Question({ question }: { question: HiddenQuestion }) {
  return (
    <div className="h-full flex justify-center items-center p-10">
      <p className="text-center text-3xl font-semibold">{question.text}</p>
    </div>
  );
}
