import React from "react";
import { BoardQuestion } from "./Room";

export default function Board({
  availableQuestions,
  chooseQuestion,
  isCurrentPlayer,
}: {
  availableQuestions: { [key: string]: BoardQuestion[] };
  chooseQuestion: (question: { category: string; index: number }) => void;
  isCurrentPlayer: boolean;
}) {
  const categoriesCount = Object.keys(availableQuestions).length;
  const questionsInCategoryCount = Object.values(availableQuestions)[0].length;
  const tableData: {
    value: number;
    hasBeenPlayed: boolean;
    onClick: () => void;
  }[][] = new Array(categoriesCount)
    .fill(undefined)
    .map(() => new Array(questionsInCategoryCount).fill(undefined));

  for (const [categoryIndex, [category, questions]] of Object.entries(
    availableQuestions
  ).entries()) {
    for (const question of questions) {
      tableData[question.index][categoryIndex] = {
        value: question.value,
        hasBeenPlayed: question.hasBeenPlayed,
        onClick: () => chooseQuestion({ category, index: question.index }),
      };
    }
  }

  return (
    <table className="w-full h-full table-fixed">
      <thead>
        <tr>
          {Object.keys(availableQuestions).map((category, index) => (
            <th className="border break-all" key={index} scope="col">
              {category}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {tableData.map((row, i) => (
          <tr key={i}>
            {row.map(({ value, hasBeenPlayed, onClick }, j) =>
              hasBeenPlayed ? (
                <td className="border"></td>
              ) : (
                <td
                  className={`text-center text-lg font-bold border${
                    isCurrentPlayer ? " hover:bg-white hover:text-black" : ""
                  }`}
                  key={j}
                  onClick={onClick}
                >
                  {value}
                </td>
              )
            )}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
